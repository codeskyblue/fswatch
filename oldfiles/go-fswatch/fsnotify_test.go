package fswatch

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWalkDir(t *testing.T) {
	Convey("WalkDir should list a directory", t, func() {
		dirs, err := WalkDir("testdata/dirs", 5)
		So(err, ShouldBeNil)
		So(len(dirs), ShouldEqual, 3)
	})

	Convey("Watch should watch add and remove", t, func() {
		evC, err := Watch([]string{"testdata"}, nil)
		So(err, ShouldBeNil)
		ioutil.WriteFile("testdata/tmp.go", []byte("haha"), 0644)
		event := <-evC
		t.Log(event)

		event = <-evC
		t.Log(event) // First create event
		So(event.Type, ShouldEqual, ET_FILESYSTEM)
		So(event.Op, ShouldEqual, Write)

		os.Chmod("testdata/tmp.go", 0755)
		event = <-evC
		So(event.Type, ShouldEqual, ET_FILESYSTEM)
		So(event.Op, ShouldEqual, Chmod)

		os.Remove("testdata/tmp.go")
		event = <-evC
		t.Log(event)
		So(event.Type, ShouldEqual, ET_FILESYSTEM)
		So(event.Op, ShouldEqual, Remove)
	})
}
