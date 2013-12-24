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

type RTMPObj interface{
    GetCurrentPositionForKey(key string) (int, error)
    SetCurrentPositionForKey(key string, position int) error
    GetRTMPScoreForKey(key string) (*RTMPScore, error)

    GetRTMPScoreFromPosition(position int) (*RTMPScore, error)
    SetRTMPScoreAtPosition(r *RTMPScore, position int) error

    GetDBName() string

    GetCursor() (int,error)
    SetCursor(position int) error

    GetMaxElements() (int, error)
    SetMaxElements(amt int) error

    DeleteKey(key string)
}

func addNewEntry(o RTMPObj, key string, value int, position int){
    key_info := RTMPScore{key, value}
    //fmt.Printf("in here loading for %s\n", key)
    o.SetCurrentPositionForKey(key, position)
    o.SetRTMPScoreAtPosition(&key_info, position)
}

func addObject(o RTMPObj, key string, value int){
    //1. find the object
    location, err := o.GetCurrentPositionForKey(key)
    if(err == nil){ //found the key
        //2. if it does exist increase its value
        rtmpObj,err := o.GetRTMPScoreFromPosition(location)
        if(err != nil){
            return
        }
        rtmpObj.score += value
        o.SetRTMPScoreAtPosition(rtmpObj, location)
    } else {
        //3. if it doesnt exist 
        //3a.    if no object exists in the current location - add this object
        location,err = o.GetCursor()
        if(err != nil){
            return
        }
        rtmpObj,_ := o.GetRTMPScoreFromPosition(location)
        if((rtmpObj == nil) || (rtmpObj.score == 0)){ //simply add the element
            addNewEntry(o, key, value, location)
        } else{ //3b.    else reduce the count of the object at the current location and if it drops to 0 - add this new object
            rtmpObj.score--
            if(rtmpObj.score == 0){
                //TODO: delete key
                o.DeleteKey(key)
                addNewEntry(o, key, value, location) //add the new entry
            } else{
                o.SetRTMPScoreAtPosition(rtmpObj, location)
            }
        }
        location++
        get_max_val, _ := o.GetMaxElements()
        location %= get_max_val
        o.SetCursor(location)
    }
}
func addObjectOne(o RTMPObj, key string){
    addObject(o, key, 1)
    return
}

func getValue(o RTMPObj, key string) (int, bool) {
    /*
    location, err := o.GetCurrentPositionForKey(key)
    if(err != nil){
        return 0, false
    }
    fmt.Printf("here3a with position = %d\n", location)

    score_obj, err := o.GetRTMPScoreFromPosition(location)
    if(err != nil){
        return 0, false
    }
    return score_obj.score, true
    */
    //fmt.Printf("name = %s", o.GetDBName())
    score_obj, err := o.GetRTMPScoreForKey(key)
    if(err != nil){
        return 0, false
    }
    return score_obj.score, true
}



/////Fetch an instance
var global_rtmp_map map[string]RTMPObj
func getRTMPObject(request string) RTMPObj{
    if(global_rtmp_map == nil){
        global_rtmp_map = make(map[string]RTMPObj)
    }

    o, ok := global_rtmp_map[request]
    if(ok == false){ //didnt find the object
        l := new(LevelDBRTMPObj)
        err := l.initLRTMPObj(request)
        if(err != nil){
            return nil
        }

        count, err := l.GetCursor()
        if(count == 0 || err != nil){
            l.SetCursor(0)
        }
        count, err = l.GetMaxElements()
        if(count == 0 || err != nil){
            l.SetMaxElements(1000000)
        }
        o = l
        global_rtmp_map[request] = o
    }

    //fmt.Printf("returning %v\n", o)
    return o
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
        split_list_len := len(split_list)
        for i:=0; i< split_list_len; i++ {
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

    collection_name, err := query_map["collection"]
    if(err == false){
        errorResponse(w, 400, "Bad Request - collection not provided")
        return
    }

    /*
    getRTMPObject(collection_name)
    fmt.Fprintf(w, `{"%s":%d}`, "x", 15)
    return;
    */

    value,_ := getValue(getRTMPObject(collection_name), key_name)
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

    collection_name, err := query_map["collection"]
    if(err == false){
        errorResponse(w, 400, "Bad Request - collection not provided")
        return
    }

    value, ok := query_map["value"]
    if(ok == true){
        amt_to_add, err_val := strconv.Atoi(value)
        if (err_val == nil){
            addObject(getRTMPObject(collection_name), key_name, amt_to_add)
        }else{
            errorResponse(w, 400, "Bad request - value incorrect")
            return
        }
        fmt.Fprintln(w, "OK")
    } else{
        errorResponse(w, 400, "Bad request - value not provided")
        return
    }
}
