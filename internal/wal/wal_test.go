package wal

import (
	"strconv"
	"testing"
	"time"
)

func TestNewWalAppend(t *testing.T) {

	fileDest, err := CreateWalFile("/tmp", "wal-")
	if err != nil {
		t.Error("wal.CreateWalFile: " + err.Error())
	}

	keys := []string{"first-test-key", "second-test-key", "third-test-key"}
	pats := make([]string, 0)

	s1 := `{
		"sys":["filestore"],
		"evt":["file-created","file-versioned"],
		"tid":["2132353463456546747456"]
	}`
	s2 := `{
		"sys":["authnz"],
		"evt":["user-login"],
		"tid":["094534435","34563445423324","456546456243423"]
	}`
	s3 := `{
		"sys":["infra"],
		"evt":["circuit-breaker-tripped","circuit-breaker-failure"]
	}`
	pats = append(pats, s1)
	pats = append(pats, s2)
	pats = append(pats, s3)

	wal, err := OpenWalFile(fileDest)
	if err != nil {
		t.Error("wal.OpenWalFile: " + err.Error())
	}
	ts := time.Now().UnixNano()
	wal.Write(ts, []byte(keys[0]), []byte(pats[0]), WAL_ADD)
	wal.Close()

	wal, err = OpenWalFile(fileDest)
	if err != nil {
		t.Error("wal.OpenWalFile: " + err.Error())
	}
	ts = time.Now().UnixNano()
	wal.Write(ts, []byte(keys[1]), []byte(pats[1]), WAL_DEL)
	wal.Close()

	wal, err = OpenWalFile(fileDest)
	if err != nil {
		t.Error("wal.OpenWalFile: " + err.Error())
	}
	ts = time.Now().UnixNano()
	wal.Write(ts, []byte(keys[2]), []byte(pats[2]), WAL_ADD)
	wal.Close()

	wal, err = OpenWalFile(fileDest)
	if err != nil {
		t.Error("wal.OpenWalFile: " + err.Error())
	}
	x := 0
	for x = 0; x < len(keys); x++ {
		if !wal.HasNext() {
			break
		}
		val := WalEntry{}
		val, err := wal.Next()

		if err != nil {
			t.Error("wal.Next: " + err.Error())
		} else {
			if string(val.Key) != keys[x] {
				t.Error("Failed to retrieve key " + strconv.Itoa(x) + " (Got *" + string(val.Key) + "*, expected *" + keys[x] + "*)")
			}
			if string(val.Pattern) != pats[x] {
				t.Error("Failed to retrieve pattern " + strconv.Itoa(x) + " (Got *" + string(val.Pattern) + "*, expected *" + pats[x] + "*)")
			}
		}
	}
	if x != 3 {
		t.Error("Not enough entries -- expected 3, got " + strconv.Itoa(x) + "!")
	}
	wal.Close()

	wal.Delete()
}

func TestNewWalFileReadWrite(t *testing.T) {
	fildest, err := CreateWalFile("/tmp/", "wal")
	if err != nil {
		t.Error("wal.CreateWalFile: " + err.Error())
	}
	wal, err := OpenWalFile(fildest)
	if err != nil {
		t.Error("wal.OpenWalFile: " + err.Error())
	}

	keys := []string{"first-test-key", "second-test-key", "third-test-key"}
	pats := make([]string, 0)

	s1 := `{
		"sys":["filestore"],
		"evt":["file-created","file-versioned"],
		"tid":["2132353463456546747456"]
	}`
	s2 := `{
		"sys":["authnz"],
		"evt":["user-login"],
		"tid":["094534435","34563445423324","456546456243423"]
	}`
	s3 := `{
		"sys":["infra"],
		"evt":["circuit-breaker-tripped","circuit-breaker-failure"]
	}`
	pats = append(pats, s1)
	pats = append(pats, s2)
	pats = append(pats, s3)

	for x := 0; x < len(keys); x++ {
		ts := time.Now().UnixNano()
		wal.Write(ts, []byte(keys[x]), []byte(pats[x]), WAL_ADD)
	}

	wal.Rewind()
	x := 0
	for x = 0; x < len(keys)+100; x++ {
		if !wal.HasNext() {
			break
		}
		val := WalEntry{}
		val, err := wal.Next()

		if err != nil {
			t.Error("wal.Next: " + err.Error())
		} else {
			if string(val.Key) != keys[x] {
				t.Error("Failed to retrieve key " + strconv.Itoa(x) + " (Got *" + string(val.Key) + "*, expected *" + keys[x] + "*)")
			}
			if string(val.Pattern) != pats[x] {
				t.Error("Failed to retrieve pattern " + strconv.Itoa(x) + " (Got *" + string(val.Pattern) + "*, expected *" + pats[x] + "*)")
			}
		}
	}
	if x != 3 {
		t.Error("Not enough entries -- expected 3, got " + strconv.Itoa(x) + "!")
	}
	wal.Delete()
}
