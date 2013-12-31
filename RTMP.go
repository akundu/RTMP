package main

import "net/http"
import "fmt"
import "log"
import "runtime"
//import "RTMP/basic_request"
import "github.com/akundu/RTMP/RTMP"
//import "os"
//import "strconv"
import "flag"


func main() {
    /*
    if(len(os.Args) > 1){
        tmp_value,err := strconv.Atoi(os.Args[1])
        if(err == nil){
            num_procs_to_run = tmp_value
        }
    }
    */
    num_procs_to_run := flag.Int("num_cpu", runtime.NumCPU() - 1, "num cpus to run on")
    port_to_run_on := flag.Int("port", 8080, "port to run on")
    ip_to_run_on := flag.String("ip", "", "ip to run on")
    flag.Parse()

    fmt.Printf("num of procs used before = %d and now putting out %d\n", runtime.GOMAXPROCS(*num_procs_to_run), *num_procs_to_run)
    http.HandleFunc("/get", RTMP.Get)
    http.HandleFunc("/add", RTMP.Add)
    //http.HandleFunc("/", basic_request.HandleRequest)
    ip_string := fmt.Sprintf("%s:%d", *ip_to_run_on, *port_to_run_on)
    fmt.Printf("listening on %s\n", ip_string)
    log.Fatal(http.ListenAndServe(ip_string, nil))
}
