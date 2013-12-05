package RTMP

import "net/http"
import "fmt"
import "strings"
import "strconv"
//import "time"



////The real guts
type RTMPScore struct{
    key     string
    score   int
}

type RTMPObj struct{
    current_position  map[string]int
    cursor            int
    max_elements      int
    lru_list          []*RTMPScore
}
func initRTMPObj() *RTMPObj{
    o := new(RTMPObj)
    o.current_position = make(map[string]int)
    o.cursor = 0
    o.lru_list = make([]*RTMPScore, 1000000)
    o.max_elements = 1000000
    return o
}

func (o *RTMPObj) addNewEntry(key string, value int){
    key_info := RTMPScore{key, value}
    o.lru_list[o.cursor] = &key_info
    o.current_position[key] = o.cursor
}

func (o *RTMPObj) addObject(key string, value int){
    //1. find the object
    location, ok := o.current_position[key]
    if(ok == true){ //found the key
        //2. if it does exist increase its value
        o.lru_list[location].score += value
    } else {
        //3. if it doesnt exist 
        //3a.    if no object exists in the current location - add this object
        if((o.lru_list[o.cursor] == nil) || (o.lru_list[o.cursor].score == 0)){ //simply add the element
            o.addNewEntry(key, value)
        } else{ //3b.    else reduce the count of the object at the current location and if it drops to 0 - add this new object
            o.lru_list[o.cursor].score--
            if(o.lru_list[o.cursor].score == 0){
                delete(o.current_position, key) //remove the existing entry
                o.addNewEntry(key, value) //add the new entry
            }
        }
        o.cursor++ //move the cursor forward
        o.cursor %= o.max_elements
    }
}
func (o *RTMPObj) addObjectOne(key string){
    o.addObject(key, 1)
    return
}

func (o *RTMPObj) getValue(key string) (int, bool) {
    location, found := o.current_position[key]
    if(found == true){ //found the key
        return o.lru_list[location].score, found
    }
    return 0,found
}



/////Fetch an instance
var global_rtmp_obj *RTMPObj
func getRTMPObject(request string) *RTMPObj{
    if(global_rtmp_obj == nil){
        global_rtmp_obj = initRTMPObj()
    }
    return global_rtmp_obj
}






/////Server (HTTP) Accessor Methods
func methodNotAllowed(w http.ResponseWriter){
    w.WriteHeader(405)
    fmt.Fprintln(w, "Method not allowed")
    return
}

func errorResponse(w http.ResponseWriter, status_code int, err_string string){
    w.WriteHeader(status_code)
    fmt.Fprintln(w, "%s", err_string)
    return
}

func fillQueryMap(r *http.Request) map[string]string{
    query_map := make(map[string]string)

    query := r.URL.RawQuery
    if(len(query) > 0){ //break up the string
        split_list := strings.Split(query, "&")
        for i:=0; i<len(split_list); i++ {
            split_value := strings.Split(split_list[i], "=")
            if(len(split_value) == 2) {
                query_map[split_value[0]] = split_value[1]
                //fmt.Printf("%s = %s\n", split_value[0], split_value[1])
            }
        }
    }
    return query_map
}

func Get(w http.ResponseWriter, r *http.Request) {
    if(r.Method != "GET"){
        methodNotAllowed(w)
        return
    }

    query_map := fillQueryMap(r)
    key_name, err := query_map["key"]
    if(err == false){
        errorResponse(w, 400, "Bad Request - key not provided")
        return
    }

    value,_ := getRTMPObject("").getValue(key_name)
    fmt.Fprintf(w, `{"%s":%d}`, key_name, value)
}

func Add(w http.ResponseWriter, r *http.Request) {
    if(r.Method != "GET"){
        methodNotAllowed(w)
        return
    }

    query_map := fillQueryMap(r)
    key_name, err := query_map["key"]
    if(err == false){
        errorResponse(w, 400, "Bad Request - key not provided")
        return
    }

    value, ok := query_map["value"]
    if(ok == true){
        amt_to_add, err_val := strconv.Atoi(value)
        if (err_val == nil){
            getRTMPObject("").addObject(key_name, amt_to_add)
        }else{
            errorResponse(w, 400, "Bad request - value incorrect")
            return
        }
        fmt.Fprintln(w, "OK")
    } else{
        errorResponse(w, 400, "Bad request - value not provided")
        return
    }

    //time.Sleep(5 * time.Second)
}
