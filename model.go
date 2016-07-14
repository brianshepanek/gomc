package gomc

import (
	"fmt"
	"encoding/json"
	//"github.com/brianshepanek/gomc/databases"
	//"gomc/config"
	"reflect"
	//"strings"
	"strconv"
	//"time"
)

type Params struct {
    Query map[string]interface{}
    Fields []string
    Sort string
    Page int
    Limit int
    //Skip int
}

type AppModel interface{
	BeforeFind(params Params)
	Find(params Params, result interface{}) (error)
	AfterFind()
	FindOne(params Params, result interface{}) (error)
	FindId(id string, result interface{}) (error)
	DeleteId(id string, result interface{}) (error)
	SetSaveAction()
	BeforeValidate()
	Validate()
	AfterValidate()
	BeforeSave()
	Save(result interface{}) (error)
	SaveBulk(result ...interface{}) (error)
	IndexBulk(docs []interface{}) (error)
	AfterSave()
	BeforeIndex()
	SaveIndex(result interface{})
	Reindex()
	FindCount(params Params) int
	DeleteIndex()
	CreateIndex()
	MapIndex()
	AfterIndex()
	BeforeCache()
	SaveCache(result interface{})
	AfterCache()
	BeforeWebSocketPush()
	WebSocketPush()
	AfterWebSocketPush()
}

type Model struct {
	AppConfig AppConfig
	UseDatabaseConfig string
	UseTable string
	Sort string
	Limit int
	PrimaryKey string
	Schema interface{}
	SaveAction string
	Data interface{}
	ValidationRules map[string][]ValidationRule
	ValidationErrors []RequestError
	IndexData bool
	IndexDataUseDatabaseConfig string
	IndexDataUseTable string
	IndexId string
	CacheData bool
	CacheDataUseDatabaseConfig string
	CacheDataUseTable string
	CacheId string
	CachePrefix string
	FoundFrom string
	WebSocketPushData bool
	WebSocketPushChannel string
	Count int
	Links []map[string]int
	IndexMapping map[string]interface{}
}


func (m *Model) SetSaveAction(){

	//Action
	if m.SaveAction == ""{

		m.SaveAction = "create"

		//Data
		sentData := m.Data
		data := reflect.ValueOf(sentData)
		if m.PrimaryKey != ""{
			primaryKey := reflect.Indirect(data).FieldByName(m.PrimaryKey)
			if !IsEmptyValue(primaryKey) {
				m.SaveAction = "update"
			}
		}
	}
}

func (m *Model) BeforeValidate(){}

func (m *Model) Validate(){
	m.ValidationErrors = make([]RequestError,0)
	m.ValidationErrors = ModelValidate(m)
}

func (m *Model) AfterValidate(){}

func Validate(am AppModel, result interface{}){

	//Set Save Action
	am.SetSaveAction()

	//Validate
	am.BeforeValidate()
	am.Validate()
	am.AfterValidate()

}

func (m *Model) BeforeSave(){}

func (m *Model) Save(result interface{}) (error){
	var err error
	m.AppConfig = Config
	if len(m.ValidationErrors) == 0 {

		//Database Config
		switch {
		case m.AppConfig.Databases[m.UseDatabaseConfig].Type == "mongodb" :

			var db MongoDb
			err := db.Save(m, result)

			if err != nil{
				fmt.Println(err)
			}
		}
	}

	return err
}

func (m *Model) AfterSave(){}

func (m *Model) BeforeIndex(){}

func (m *Model) SaveIndex(result interface{}){
	m.AppConfig = Config
	if m.IndexData {

		if len(m.ValidationErrors) == 0 {

			//Database Config
			switch {
			case m.AppConfig.Databases[m.IndexDataUseDatabaseConfig].Type == "elasticsearch" :
				var db Elasticsearch
				err := db.Index(m, result)
				if err != nil{

				}
			}
		}
	}
}

func Reindex(am AppModel, params Params, results interface{}){
	am.DeleteIndex()
	am.CreateIndex()
	am.MapIndex()
}

func FindCount(am AppModel, params Params) int{
	count := am.FindCount(params)
	return count
}

func (m *Model) FindCount(params Params) int{
	var count int
	m.AppConfig = Config
	if m.IndexData {

		switch {
		case m.AppConfig.Databases[m.UseDatabaseConfig].Type == "mongodb" :
			var db MongoDb
			count = db.Count(m, params)
		}
	}
	return count
}

func (m *Model) Reindex(){}

func (m *Model) CreateIndex(){
	m.AppConfig = Config
	if m.IndexData {

		//Database Config
		switch {
		case m.AppConfig.Databases[m.IndexDataUseDatabaseConfig].Type == "elasticsearch" :
			var db Elasticsearch
			err := db.CreateIndex(m)
			if err != nil{

			}
		}
	}
}

func (m *Model) DeleteIndex(){
	m.AppConfig = Config
	if m.IndexData {

		//Database Config
		switch {
		case m.AppConfig.Databases[m.IndexDataUseDatabaseConfig].Type == "elasticsearch" :
			var db Elasticsearch
			err := db.DeleteIndex(m)
			if err != nil{

			}
		}
	}
}

func (m *Model) MapIndex(){
	m.AppConfig = Config
	if m.IndexData {

		//Database Config
		switch {
		case m.AppConfig.Databases[m.IndexDataUseDatabaseConfig].Type == "elasticsearch" :
			var db Elasticsearch
			err := db.MapIndex(m)
			if err != nil{

			}
		}
	}
}

func (m *Model) AfterIndex(){}

func (m *Model) BeforeCache(){}

func (m *Model) SaveCache(result interface{}){
	m.AppConfig = Config
	if m.CacheData {

		if len(m.ValidationErrors) == 0 {

			//Database Config
			switch {
			case m.AppConfig.Databases[m.CacheDataUseDatabaseConfig].Type == "redis" :
				var db Redis
				err := db.Set(m, result)
				if err != nil{

				}
			}
		}
	}
}

func (m *Model) AfterCache(){}

func (m *Model) BeforeWebSocketPush(){}

func (m *Model) WebSocketPush(){
	m.AppConfig = Config
	if m.WebSocketPushData {

		if len(m.ValidationErrors) == 0 {

			jsonData, _ := json.Marshal(m.Data)
			err := WebSocketPush(m.WebSocketPushChannel, string(jsonData))
			if err != nil {

			}
		}
	}
}

func (m *Model) AfterWebSocketPush(){}

func IndexBulk(am AppModel, docs []interface{}) (error){

	var err error
	err = am.IndexBulk(docs)
	return err
}

func (m *Model) IndexBulk(docs []interface{}) (error){
	var err error
	if m.IndexData {

		//Database Config
		switch {
		case m.AppConfig.Databases[m.IndexDataUseDatabaseConfig].Type == "elasticsearch" :
			var db Elasticsearch
			err := db.IndexBulk(m, docs)
			if err != nil{

			}
		}
	}

	return err
}


func SaveBulk(am AppModel, records ...interface{}) (error){

	var err error
	err = am.SaveBulk(records...)
	return err
}

func (m *Model) SaveBulk(records ...interface{}) (error){
	var err error
	m.AppConfig = Config
	//Database Config
	switch {
	case m.AppConfig.Databases[m.UseDatabaseConfig].Type == "mongodb" :

		var db MongoDb
		err := db.SaveBulk(m, records...)
		if err != nil{
			//fmt.Println(err)
		}
	}

	return err
}

func Save(am AppModel, result interface{}) (error){

	var err error

	//Set Save Action
	am.SetSaveAction()

	//Validate
	am.BeforeValidate()
	am.Validate()
	am.AfterValidate()

	//Save
	am.BeforeSave()
	err = am.Save(result)
	am.AfterSave()

	//Index
	am.BeforeIndex()
	am.SaveIndex(result)
	am.AfterIndex()

	//Cache
	am.BeforeCache()
	am.SaveCache(result)
	am.AfterCache()

	//WebSocket Push
	am.BeforeWebSocketPush()
	am.WebSocketPush()
	am.AfterWebSocketPush()

	return err
}


func (m *Model) FindId(id string, result interface{}) (error){

	var err error
	m.AppConfig = Config
	if m.CacheData {
		switch {
		case m.AppConfig.Databases[m.CacheDataUseDatabaseConfig].Type == "redis" :
			var db Redis
			err = db.Get(m, m.CachePrefix + id, result)

		}
	} else {
		switch {
		case m.AppConfig.Databases[m.UseDatabaseConfig].Type == "mongodb" :

			var db MongoDb
			err := db.FindId(m, id, result)
			if err != nil{

			}
		}
	}

	return err
}

func FindId(am AppModel, id string, result interface{}) (error){

	var err error

	//Database Config
	err = am.FindId(id, result)


	return err
}

func (m *Model) FindOne(params Params, result interface{}) (error){

	var err error
	m.AppConfig = Config
	switch {
	case m.AppConfig.Databases[m.UseDatabaseConfig].Type == "mongodb" :

		var db MongoDb
		err = db.FindOne(m, params, result)

	}

	return err
}

func FindOne(am AppModel, params Params, result interface{}) (error){
	var err error

	//Database Config
	err = am.FindOne(params, result)


	return err
}

func (m *Model) Find(params Params, results interface{}) (error){

	var err error
	m.AppConfig = Config
	if m.IndexData {
		switch {
		case m.AppConfig.Databases[m.IndexDataUseDatabaseConfig].Type == "elasticsearch" :

			var db Elasticsearch
			err = db.Query(m, params, results)

		}
	} else {
		switch {
		case m.AppConfig.Databases[m.UseDatabaseConfig].Type == "mongodb" :

			var db MongoDb
			err := db.Find(m, params, results)
			if err != nil{

			}

		}
	}

	//First
	if m.Count > m.Limit {
		x := 1
		var pages []int
		for i := 1; i < m.Count; i += m.Limit{
		    pages = append(pages, x)
			x++
		}
		if len(pages) > 0 {
			m.Links = append(m.Links, map[string]int{"first" : 1})
			m.Links = append(m.Links, map[string]int{"last" : pages[len(pages) - 1]})
			if params.Page - 1 > 0 {
				m.Links = append(m.Links, map[string]int{"prev" : params.Page - 1})
			}
			if params.Page + 1 <= len(pages) - 1 {
				m.Links = append(m.Links, map[string]int{"next" : params.Page + 1})
			}
			for y := 1; y <= len(pages); y++{
				m.Links = append(m.Links, map[string]int{"page_" + strconv.Itoa(y) : y})
			}
		}
	}
	return err
}

func Find(am AppModel, params Params, results interface{}) (error){
	var err error

	//Validate
	am.BeforeFind(params)
	err = am.Find(params, results)
	am.AfterFind()

	return err
}

func (m *Model) BeforeFind(params Params){}

func (m *Model) AfterFind(){}


func (m *Model) DeleteId(id string, result interface{}) (error){

	var err error
	m.AppConfig = Config
	if m.IndexData {
		switch {
		case m.AppConfig.Databases[m.IndexDataUseDatabaseConfig].Type == "elasticsearch" :

			var db Elasticsearch
			err = db.Delete(m, id, result)

		}
	}

	if m.CacheData {
		switch {
		case m.AppConfig.Databases[m.CacheDataUseDatabaseConfig].Type == "redis" :

			var db Redis
			err = db.Delete(m, m.CachePrefix + id, result)

		}
	}

	switch {
	case m.AppConfig.Databases[m.UseDatabaseConfig].Type == "mongodb" :

		var db MongoDb
		err := db.Delete(m, id, result)
		if err != nil{

		}
	}
	return err
}

func DeleteId(am AppModel, id string, result interface{}) (error){

	var err error

	//Database Config
	err = am.DeleteId(id, result)


	return err
}