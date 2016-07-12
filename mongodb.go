package gomc

import (
	"crypto/tls"
  	"net"
	"fmt"
	"reflect"
	"gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    //"time"
)

type MongoDb struct {}

type MongoDbParams struct {
	Query map[string]interface{}
    Fields map[string]int
    Sort string
    Limit int
    Skip int
}


func (db MongoDb) Connect(server string) (*mgo.Session, error) {

	tlsConfig := &tls.Config{}
  	tlsConfig.InsecureSkipVerify = true

	//fmt.Println(server)
	dialInfo, err := mgo.ParseURL(server)
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, err
	}

	//Here it is the session. Up to you from here ;)
	session, err := mgo.DialWithInfo(dialInfo)
    //session, err := mgo.Dial(server)
    if err != nil {
    	fmt.Println(err)
    }
    //defer session.Close()

    return session, err
}


func (db MongoDb) Collection(host string, database string, collection string) (*mgo.Collection, *mgo.Session) {

	sess, err := db.Connect(host)
	if err != nil {
    	//panic(err)
    }
	coll := sess.DB(database).C(collection)

    return coll, sess
}

func (db MongoDb) formatParams(model *Model, params Params) (MongoDbParams){

	//Mongo Params
	var mongoDbParams MongoDbParams

	//Fields
	fields := make(map[string]int)
	for _, element := range params.Fields{
		fields[element] = 1
	}
	mongoDbParams.Fields = fields

	//Sort
	mongoDbParams.Sort = model.Sort
	if params.Sort != "" {
		mongoDbParams.Sort = params.Sort
	}

	//Page
	mongoDbParams.Skip = 0
	mongoDbParams.Limit = model.Limit
	if params.Limit != 0{
		mongoDbParams.Limit = params.Limit
	}
	if params.Page > 0 {
		mongoDbParams.Skip = ((params.Page - 1) * model.Limit)
		mongoDbParams.Limit = (params.Page * model.Limit)
	}
	for key, value := range params.Query{
		structKey := StructKeyFromJsonKey(model.Schema, key)
		//fmt.Println(structKey)
		field, err := reflect.TypeOf(model.Schema).FieldByName(structKey)
		if err {
			fieldType := field.Type.Kind()

			switch fieldType {
			case reflect.String, reflect.Array:
				//params.Query[key] = valueField.String()
			case reflect.Bool:
				if(value == "true" || value == 1 || value == "1"){
					params.Query[key] = true
				} else {
					params.Query[key] = false
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				//params.Query[key] = valueField.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				//params.Query[key] = valueField.Uint()
			case reflect.Float32, reflect.Float64:
				//params.Query[key] = valueField.Float()
			default :
				//params.Query[key] = valueField.Interface()
			}
		}
	}

	//Query
	mongoDbParams.Query = params.Query
	return mongoDbParams
}

func (db MongoDb) Delete(model *Model, id string, result interface{}) error {

	//DB
    collection, session := db.Collection(model.AppConfig.Databases[model.UseDatabaseConfig].Host, model.AppConfig.Databases[model.UseDatabaseConfig].Database, model.UseTable)
    defer session.Close()

    collectionDelete, sessionDelete := db.Collection(model.AppConfig.Databases[model.UseDatabaseConfig].Host, model.AppConfig.Databases[model.UseDatabaseConfig].Database, "_" + model.UseTable)
    defer sessionDelete.Close()

    var err error

    //Find Current
    err = collection.FindId(bson.ObjectIdHex(id)).One(result)
	if err != nil {
		//fmt.Println(err)
	}

	//Move to Delete Collection
	err = collectionDelete.Insert(result)

	//Delete Current
	err = collection.RemoveId(bson.ObjectIdHex(id))
	fmt.Println(err)
	return err
}

func (db MongoDb) Save(model *Model, result interface{}) error {

    //DB
    collection, session := db.Collection(model.AppConfig.Databases[model.UseDatabaseConfig].Host, model.AppConfig.Databases[model.UseDatabaseConfig].Database, model.UseTable)
    defer session.Close()

    //Results
   	var err error

   	//Created
   	if model.SaveAction == "create" {
    	err = collection.Insert(model.Data)
    	primaryKey := reflect.ValueOf(model.Data).FieldByName(model.PrimaryKey).Interface()
    	err = collection.FindId(primaryKey).One(result)
    	if err != nil {
    		fmt.Println(err)
    	}
	}

	//Replace
	if model.SaveAction == "replace" {
		primaryKey := reflect.ValueOf(model.Data).FieldByName(model.PrimaryKey).Interface()
		err = collection.Update(bson.M{"_id": primaryKey}, model.Data)
		if err != nil {
    		fmt.Println(err)
    	}
	}

	//Updated
	if model.SaveAction == "update" {

		//Update Map
		updateMap := make(map[string]map[string]interface{})
    	updateMap["$set"] = make(map[string]interface{})
    	data := reflect.ValueOf(model.Data)
    	primaryKey := reflect.ValueOf(model.Data).FieldByName(model.PrimaryKey).Interface()

    	for i := 0; i < data.NumField(); i++ {
			valueField := data.Field(i)
			nameField := data.Type().Field(i).Name
			if !IsEmptyValue(valueField) && nameField != model.PrimaryKey{
				switch valueField.Kind() {
				case reflect.String, reflect.Array:
					updateMap["$set"][JsonKeyFromStructKey(model.Schema, nameField)] = valueField.String()
				case reflect.Bool:
					updateMap["$set"][JsonKeyFromStructKey(model.Schema, nameField)] = valueField.Bool()
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					updateMap["$set"][JsonKeyFromStructKey(model.Schema, nameField)] = valueField.Int()
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					updateMap["$set"][JsonKeyFromStructKey(model.Schema, nameField)] = valueField.Uint()
				case reflect.Float32, reflect.Float64:
					updateMap["$set"][JsonKeyFromStructKey(model.Schema, nameField)] = valueField.Float()
				default :
					updateMap["$set"][JsonKeyFromStructKey(model.Schema, nameField)] = valueField.Interface()
				}
			}
		}
		change := mgo.Change{
			Update: updateMap,
			ReturnNew: true,
		}
		_, err = collection.Find(bson.M{"_id": primaryKey}).Apply(change, result)

	}
    return err
}

func (db MongoDb) FindId(model *Model, id string, result interface{}) error {

	//DB
	collection, session := db.Collection(model.AppConfig.Databases[model.UseDatabaseConfig].Host, model.AppConfig.Databases[model.UseDatabaseConfig].Database, model.UseTable)
    defer session.Close()

    //Results
   	err := collection.FindId(bson.ObjectIdHex(id)).One(result)

	return err
}

func (db MongoDb) FindOne(model *Model, params Params, result interface{}) error {

	//Params
	mongoParams := db.formatParams(model, params)

	//DB
	collection, session := db.Collection(model.AppConfig.Databases[model.UseDatabaseConfig].Host, model.AppConfig.Databases[model.UseDatabaseConfig].Database, model.UseTable)
    defer session.Close()

    //Results
   	err := collection.Find(mongoParams.Query).Select(mongoParams.Fields).Sort(mongoParams.Sort).Limit(mongoParams.Limit).Skip(mongoParams.Skip).One(result)

	return err
}

func (db MongoDb) Count(model *Model) int {

	//DB
	collection, session := db.Collection(model.AppConfig.Databases[model.UseDatabaseConfig].Host, model.AppConfig.Databases[model.UseDatabaseConfig].Database, model.UseTable)
    defer session.Close()

	//Count
	countResult, countErr := collection.Count()

	if countErr != nil {
	    // Handle error
	    fmt.Println(countErr)
	} else {
		model.Count = countResult
		//count = countResult
	}

	return countResult
}

func (db MongoDb) Find(model *Model, params Params, results interface{}) error {

	//Params
	mongoParams := db.formatParams(model, params)

	//fmt.Println(mongoParams)

	//DB
	collection, session := db.Collection(model.AppConfig.Databases[model.UseDatabaseConfig].Host, model.AppConfig.Databases[model.UseDatabaseConfig].Database, model.UseTable)
    defer session.Close()

	//Count
	countResult, countErr := collection.Find(mongoParams.Query).Select(mongoParams.Fields).Sort(mongoParams.Sort).Count()

	if countErr != nil {
	    // Handle error
	    fmt.Println(countErr)
	} else {
		model.Count = countResult
		//count = countResult
	}

    //Results
   	err := collection.Find(mongoParams.Query).Select(mongoParams.Fields).Sort(mongoParams.Sort).Limit(mongoParams.Limit).Skip(mongoParams.Skip).All(results)

	return err
}