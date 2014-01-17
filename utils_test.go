package main

import "testing"

func TestMd5sum(t *testing.T) {
	sum := Md5sum([]byte("hello\n"))
	t.Log(sum)
	if sum != "b1946ac92492d2347c6235b4d2611184" {
		t.Errorf("write sum failed, got: %s", sum)
	}
}

func TestMd5sumFile(t *testing.T) {
	sum, err := Md5sumFile("testdata/md5file.txt")
	if err != nil {
		t.Error(err)
	}
	if sum != "b1946ac92492d2347c6235b4d2611184" {
		t.Errorf("write sum failed, got: %s", sum)
	}
}
