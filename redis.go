package gomc

import (
	"fmt"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

type Redis struct {}

func (db Redis) Set(model *Model, result interface{}) error{
	
    //Connect
	conn, err := redis.Dial("tcp", model.AppConfig.Databases[model.CacheDataUseDatabaseConfig].Host + ":" + model.AppConfig.Databases[model.CacheDataUseDatabaseConfig].Port)
    if err != nil {
        fmt.Println(err)
    }

    //Key
    key := model.AppConfig.Databases[model.CacheDataUseDatabaseConfig].Database + ":" + model.CacheDataUseTable + ":" + model.CacheId
    
    jsonData, _ := json.Marshal(result)
    conn.Do("SET", key, jsonData)
    
	return err
}

func (db Redis) Get(model *Model, id string, result interface{}) error{
	
	//Redis Connect
    conn, err := redis.Dial("tcp", model.AppConfig.Databases[model.CacheDataUseDatabaseConfig].Host + ":" + model.AppConfig.Databases[model.CacheDataUseDatabaseConfig].Port)
    if err != nil {

    }

    //Key
    key := model.AppConfig.Databases[model.CacheDataUseDatabaseConfig].Database + ":" + model.CacheDataUseTable + ":" + id
    
    jsonData, _ := redis.Bytes(conn.Do("GET", key))

    err = json.Unmarshal(jsonData, result)
    model.FoundFrom = "redis"
    
    return err
}

func (db Redis) Delete(model *Model, id string, result interface{}) error{
    
    //Redis Connect
    conn, err := redis.Dial("tcp", model.AppConfig.Databases[model.CacheDataUseDatabaseConfig].Host + ":" + model.AppConfig.Databases[model.CacheDataUseDatabaseConfig].Port)
    if err != nil {

    }

    //Key
    key := model.AppConfig.Databases[model.CacheDataUseDatabaseConfig].Database + ":" + model.CacheDataUseTable + ":" + id
    
    conn.Do("DEL", key)

    return err
}