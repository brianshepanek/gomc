package gomc

import (
	"strings"
	"reflect"
	"strconv"
    "crypto/rand"
    "encoding/base64"
    "crypto/sha256"
    "io"
    "encoding/hex"
)

func IsEmptyValue(v reflect.Value) bool {
    switch v.Kind() {
    case reflect.String, reflect.Array:
        return v.Len() == 0
    case reflect.Map, reflect.Slice:
        return v.Len() == 0 || v.IsNil()
    case reflect.Bool:
        return !v.Bool()
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        return v.Int() == 0
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
        return v.Uint() == 0
    case reflect.Float32, reflect.Float64:
        return v.Float() == 0
    case reflect.Interface, reflect.Ptr:
        return v.IsNil()
    }

    return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}



func StringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func UrlMapToParams(urlMap map[string][]string) (Params){
    
    //Params
    var params Params
    
    queryMap := make(map[string]interface{})

    //Token
    if(len(urlMap["token"]) > 0){
        delete(urlMap, "token")
    }

    //Fields
    if(len(urlMap["fields"]) > 0){
        params.Fields = strings.Split(urlMap["fields"][0], ",")
        delete(urlMap, "fields")
    }

    //Sort
    if(len(urlMap["sort"]) > 0){
        params.Sort = urlMap["sort"][0]
        delete(urlMap, "sort")
    }

    //Page
    if(len(urlMap["page"]) > 0){
        page := urlMap["page"][0]
        i, _ := strconv.Atoi(page)
        params.Page = i
        delete(urlMap, "page")
    }

    //Query
    for key, value := range urlMap {
        queryMap[key] = value[0]
    }
    params.Query = queryMap
    
    return params
}

func JsonKeyFromStructKey(model interface{}, structKey string) (jsonKey string){
	field, _ := reflect.TypeOf(model).FieldByName(structKey)
    tag := string(field.Tag.Get("json"))
    tagArray := strings.Split(tag,",")

    return tagArray[0]
}

func GenerateRandomBytes(n int) ([]byte, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    // Note that err == nil only if we read len(b) bytes.
    if err != nil {
        return nil, err
    }

    return b, nil
}

func GenerateRandomString(s int) (string, error) {
    b, err := GenerateRandomBytes(s)
    return base64.URLEncoding.EncodeToString(b), err
}

func HashString(salt string, input string) (string){

    h256 := sha256.New()
    io.WriteString(h256, salt + input)
    hashedString := hex.EncodeToString(h256.Sum(nil))

    return hashedString
}