package ddns

import (
	"errors"
	"testing"

	"github.com/mkxxq/ddns/utils"
)

type fakeGetter struct {
	ip  string
	err error
}

func (g fakeGetter) GetOuterIP() (string, error) { return g.ip, g.err }

type fakeRecorder struct {
	calls []string
	err   error
}

func (r *fakeRecorder) UpsertRecord(domain string, ip string) error {
	r.calls = append(r.calls, ip)
	return r.err
}

func fixedLookup(ip string) func(string) (string, error) {
	return func(string) (string, error) { return ip, nil }
}

// 记录与出口 IP 不一致时必须落盘, 并回到稳态间隔。
func TestWatcherRunUpdatesStaleRecord(t *testing.T) {
	rec := &fakeRecorder{}
	w := Watcher{
		Domain:   "cloud.example.com",
		IPGetter: fakeGetter{ip: "171.212.137.61"},
		lookup:   fixedLookup("171.212.137.1"),
	}

	wait := w.Run(rec)

	if len(rec.calls) != 1 || rec.calls[0] != "171.212.137.61" {
		t.Fatalf("unexpected upsert calls: %#v", rec.calls)
	}
	if wait != steadyInterval {
		t.Fatalf("wait = %v, want %v", wait, steadyInterval)
	}
}

// 记录已经正确时不得写 DNS。
func TestWatcherRunNoopWhenRecordMatches(t *testing.T) {
	rec := &fakeRecorder{}
	w := Watcher{
		Domain:   "cloud.example.com",
		IPGetter: fakeGetter{ip: "171.212.137.61"},
		lookup:   fixedLookup("171.212.137.61"),
	}

	wait := w.Run(rec)

	if len(rec.calls) != 0 {
		t.Fatalf("unexpected upsert calls: %#v", rec.calls)
	}
	if wait != steadyInterval {
		t.Fatalf("wait = %v, want %v", wait, steadyInterval)
	}
}

// 拿不到出口 IP 时必须短间隔重试, 而不是等下个整分钟。
func TestWatcherRunRetriesAfterGetterError(t *testing.T) {
	rec := &fakeRecorder{}
	w := Watcher{
		Domain:   "cloud.example.com",
		IPGetter: fakeGetter{err: errors.New("i/o timeout")},
		lookup:   fixedLookup("171.212.137.1"),
	}

	if wait := w.Run(rec); wait != retryInterval {
		t.Fatalf("wait = %v, want %v", wait, retryInterval)
	}
	if len(rec.calls) != 0 {
		t.Fatalf("unexpected upsert calls: %#v", rec.calls)
	}
}

// 写 DNS 失败时也必须短间隔重试。
func TestWatcherRunRetriesAfterUpsertError(t *testing.T) {
	rec := &fakeRecorder{err: errors.New("cloudflare api error")}
	w := Watcher{
		Domain:   "cloud.example.com",
		IPGetter: fakeGetter{ip: "171.212.137.61"},
		lookup:   fixedLookup("171.212.137.1"),
	}

	if wait := w.Run(rec); wait != retryInterval {
		t.Fatalf("wait = %v, want %v", wait, retryInterval)
	}
}

// 域名查询失败不应阻塞更新: 缓存为空时按"需要修改"处理并落盘。
func TestWatcherRunContinuesWhenLookupFails(t *testing.T) {
	rec := &fakeRecorder{}
	w := Watcher{
		Domain:   "cloud.example.com",
		IPGetter: fakeGetter{ip: "171.212.137.61"},
		lookup: func(string) (string, error) {
			return "", errors.New("i/o timeout")
		},
	}

	if wait := w.Run(rec); wait != steadyInterval {
		t.Fatalf("wait = %v, want %v", wait, steadyInterval)
	}
	if len(rec.calls) != 1 || rec.calls[0] != "171.212.137.61" {
		t.Fatalf("unexpected upsert calls: %#v", rec.calls)
	}
}

// 落盘成功后, 下一轮读到的还是 TTL 内的旧记录, 不得因此再写一次。
func TestWatcherRunDoesNotRewriteWithinTTL(t *testing.T) {
	rec := &fakeRecorder{}
	w := Watcher{
		Domain:   "cloud.example.com",
		IPGetter: fakeGetter{ip: "171.212.137.61"},
		lookup:   fixedLookup("171.212.137.1"),
	}

	w.Run(rec)
	if wait := w.Run(rec); wait != steadyInterval {
		t.Fatalf("second run wait = %v, want %v", wait, steadyInterval)
	}
	if len(rec.calls) != 1 {
		t.Fatalf("record rewritten within TTL: %#v", rec.calls)
	}
}

var _ utils.OuterIPGetter = fakeGetter{}
