package main

import "net/http"
import "fmt"
import "log"
import "runtime"
//import "RTMP/basic_request"
import "github.com/akundu/RTMP/RTMP"
import "os"
import "strconv"

//import "time"

func main() {
    num_procs_to_run := 4
    if(len(os.Args) > 1){
        tmp_value,err := strconv.Atoi(os.Args[1])
        if(err == nil){
            num_procs_to_run = tmp_value
        }
    }

    fmt.Printf("num of procs used before = %d and now putting out %d\n", runtime.GOMAXPROCS(num_procs_to_run), num_procs_to_run)
    http.HandleFunc("/get", RTMP.Get)
    http.HandleFunc("/add", RTMP.Add)
    //http.HandleFunc("/", basic_request.HandleRequest)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
