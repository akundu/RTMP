package RTMP

import "strconv"
//import "fmt"
import "github.com/jmhodges/levigo"

////The real guts
type LevelDBRTMPObj struct{
    name   string
    db     *levigo.DB
    ro     *levigo.ReadOptions
    wo     *levigo.WriteOptions
}

func (o* LevelDBRTMPObj) initLRTMPObj(name string) error{
    opts := levigo.NewOptions()
    opts.SetCache(levigo.NewLRUCache(3<<30))
    opts.SetCreateIfMissing(true)

    o.name = "/tmp/" + name
    db, err := levigo.Open(o.name, opts)
    if(err != nil){
        return err
    }
    o.db = db
    o.wo = levigo.NewWriteOptions()
    o.ro = levigo.NewReadOptions()

    //fmt.Printf("db set")
    return nil
}

func (o* LevelDBRTMPObj) writeKey(key []byte, value []byte) error{
    //fmt.Printf("setting %s to %s \n", string(key), string(value));
    return o.db.Put(o.wo, key, value)
}

func (o* LevelDBRTMPObj) readKey(key []byte) ([]byte, error){
    data, err := o.db.Get(o.ro, key)
    if(err != nil){
        return nil, err
    }
    //fmt.Printf("got data = %s for %s\n", string(data), string(key))
    return data,nil
}

const CURRENT_POS_STRING = "-c-p-"
func (o* LevelDBRTMPObj) GetCurrentPositionForKey(key string) (int, error){
    position_string := string(CURRENT_POS_STRING + key)
    //fmt.Printf("looking at position string %s\n", position_string)
    data_str,err := o.readKey([]byte(position_string))
    if(err != nil){
        return 0,err
    }
    //fmt.Printf("got position for %s to be at %s\n", key, string(data_str))
    return strconv.Atoi(string(data_str))
}
func (o* LevelDBRTMPObj) SetCurrentPositionForKey(key string, position int) error{
    position_string := string(CURRENT_POS_STRING + key)
    return o.writeKey([]byte(position_string), []byte(strconv.FormatInt(int64(position), 10)))
}

const CURSOR_STRING = "-cursor-"
func (o* LevelDBRTMPObj) GetCursor() (int, error){
    data_str,err := o.readKey([]byte(CURSOR_STRING))
    if(err != nil){
        return 0,err
    }
    return strconv.Atoi(string(data_str))
}
func (o* LevelDBRTMPObj) SetCursor(position int) error{
    return o.writeKey([]byte(CURSOR_STRING), []byte(strconv.FormatInt(int64(position), 10)))
}

const MAX_ELEMENTS = "-m-e-"
func (o* LevelDBRTMPObj) GetMaxElements() (int, error){
    data_str,err := o.readKey([]byte(MAX_ELEMENTS))
    if(err != nil){
        return 0,err
    }
    return strconv.Atoi(string(data_str))
}
func (o* LevelDBRTMPObj) SetMaxElements(amt int) error{
    return o.writeKey([]byte(MAX_ELEMENTS), []byte(strconv.FormatInt(int64(amt), 10)))
}

const POSITION_ELEMENT_STRING = "-p-e-s-"
func (o* LevelDBRTMPObj) GetRTMPScoreFromPosition(position int) (*RTMPScore, error){
    //get the key at the position
    key := string(POSITION_ELEMENT_STRING + strconv.FormatInt(int64(position), 10))
    data_str,err := o.readKey([]byte(key))
    if(err != nil){
        return nil, err
    }
    return o.GetRTMPScoreForKey(string(data_str))
}

const SCORE_STRING = "-score-"
func (o* LevelDBRTMPObj) GetRTMPScoreForKey(key string) (*RTMPScore, error){
    value_key := SCORE_STRING + key
    data_str,err := o.readKey([]byte(value_key))
    if(err != nil){
        //fmt.Printf("didnt find key %s\n", key)
        return nil, err
    }
    value_of_key,err := strconv.Atoi(string(data_str))
    if(err != nil){
        //fmt.Printf("couldnt atoi value %s\n", string(data_str))
        return nil, err
    }

    obj_to_return := RTMPScore{key, value_of_key}
    return &obj_to_return, nil
}

func (o* LevelDBRTMPObj) SetRTMPScoreAtPosition(obj *RTMPScore, position int) error{
    key := POSITION_ELEMENT_STRING + strconv.FormatInt(int64(position), 10)
    err := o.writeKey([]byte(key), []byte(obj.key))
    if(err != nil){
        return err
    }

    value_key := SCORE_STRING + obj.key
    return o.writeKey([]byte(value_key), []byte(strconv.FormatInt(int64(obj.score), 10)))
}

func (o* LevelDBRTMPObj) DeleteKey(key string){
    wo := levigo.NewWriteOptions()
    o.db.Delete(wo, []byte(key))
    return
}

func (o* LevelDBRTMPObj) GetDBName() string{
    return o.name
}
