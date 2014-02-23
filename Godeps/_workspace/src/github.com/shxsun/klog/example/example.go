package main
import "github.com/shxsun/klog"
func main(){
	k := klog.NewLogger(nil, "") // Write to stdout and without prefix

	k.Infof("Hi %s.", "Susan")
	k.Warn("Oh my god, you are alive!")
	k.Error("Yes, but I will go to Mars tomorrow. So only one day with you")
	k.Fatal("Oh no, donot leave me again... faint") // Fatal will call os.Exit(1)
}
