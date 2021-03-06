package gomc

import (
	"crypto/tls"
  	"net"
	"fmt"
	"reflect"
	"gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
	"strings"
    //"time"
)

type MongoDb struct {}

type MongoDbParams struct {
	Query bson.M
	RegexQuery []interface{}
    Fields map[string]int
    Sort string
    Limit int
    Skip int
}

type Counter struct {
	Id string `bson:"id" json:"id"`
	Seq int `bson:"seq" json:"seq"`
}

func CreateMongoSession(db DatabaseConfig) (error, *mgo.Session){

	host := db.Host

	tlsConfig := &tls.Config{}
  	tlsConfig.InsecureSkipVerify = true

	//fmt.Println(server)
	dialInfo, err := mgo.ParseURL(host)

	if db.UseSSL {
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}
	}

	if db.Username != "" {
		dialInfo.Username = db.Username
	}

	if db.Password != "" {
		dialInfo.Password = db.Password
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
    	fmt.Println(err)
    }

	return err, session
}

func (db MongoDb) Collection(model *Model, counter bool) (*mgo.Collection, *mgo.Session) {

	mongoSession := Databases[model.UseDatabaseConfig].MongoSession
	databaseName := Databases[model.UseDatabaseConfig].Database
	var collectionName string
	if counter {
		collectionName = model.UseTableCounter
	} else {
		collectionName = model.UseTable
	}


	sessionCopy := mongoSession.Copy()
	collection := sessionCopy.DB(databaseName).C(collectionName)

    return collection, sessionCopy
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
	}

	//Query
	for key, value := range params.Query{
		structKey := StructKeyFromJsonKey(model.Schema, key)
		field, err := reflect.TypeOf(model.Schema).FieldByName(structKey)
		if err {
			fieldType := field.Type.Kind()
			switch fieldType {
			case reflect.String:
				if strings.Contains(value.(string), "*"){
					params.Query[key] = bson.RegEx{strings.Replace(value.(string), "*", ".*", 1), "i"}
				}
			case reflect.Array:
				//params.Query[key] = valueField.String()
			case reflect.Bool:
				if(value == "true" || value == 1 || value == "1" || value == true){
					params.Query[key] = true
				} else {
					params.Query[key] = false
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			case reflect.Float32, reflect.Float64:
			default :
			}
		}
	}

	//Query
	mongoDbParams.Query = params.Query
	return mongoDbParams
}

func (db MongoDb) Delete(model *Model, id string, result interface{}) error {

	//DB
    collection, session := db.Collection(model, false)
    defer session.Close()

    collectionDelete, sessionDelete := db.Collection(model, false)
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

func (db MongoDb) SaveBulk(model *Model, docs ...interface{}) error {



	//DB
	collection, session := db.Collection(model, false)
	defer session.Close()

	bulk := collection.Bulk()
	bulk.Upsert(docs...)

	_, err := bulk.Run()
	if err != nil {
		fmt.Println(err)
	}
	//Debug(docs)
	return err
}

func (db MongoDb) Save(model *Model, result interface{}) error {


    //DB
    collection, session := db.Collection(model, false)


    defer session.Close()

    //Results
   	var err error

   	//Created
   	if model.SaveAction == "create" {
    	err = collection.Insert(model.Data)
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
	collection, session := db.Collection(model, false)
    defer session.Close()

    //Results
   	err := collection.FindId(bson.ObjectIdHex(id)).One(result)

	return err
}

func (db MongoDb) FindOne(model *Model, params Params, result interface{}) error {

	//Params
	mongoParams := db.formatParams(model, params)

	//DB
	collection, session := db.Collection(model, false)
    defer session.Close()

    //Results
   	err := collection.Find(mongoParams.Query).Select(mongoParams.Fields).Sort(mongoParams.Sort).Limit(mongoParams.Limit).Skip(mongoParams.Skip).One(result)

	return err
}

func (db MongoDb) Counter(model *Model) int {

	var doc Counter

	//DB
	collection, session := db.Collection(model, true)
    defer session.Close()

	change := mgo.Change{
        Update: bson.M{"$inc": bson.M{"seq": 1}},
        ReturnNew: true,
	}
	_, err := collection.Find(bson.M{"_id": model.CounterId}).Apply(change, &doc)
	if err != nil {
		fmt.Println(err)
	}

	return doc.Seq
}

func (db MongoDb) Count(model *Model, params Params) int {

	mongoParams := db.formatParams(model, params)

	//DB
	collection, session := db.Collection(model, false)
    defer session.Close()

	//Count
	countResult, countErr := collection.Find(mongoParams.Query).Count()

	if countErr != nil {
	    // Handle error
	    fmt.Println(countErr)
	} else {
		model.Count = countResult
		//count = countResult
	}

	return countResult
}

func (db MongoDb) FindAggregate(model *Model, aggregate interface{}, results interface{}) error {

	//DB
	collection, session := db.Collection(model, false)
	defer session.Close()

	/*
	//Count
	countResult, countErr := collection.Pipe(aggregate).Count()

	if countErr != nil {
	    // Handle error
	    fmt.Println(countErr)
	} else {
		model.Count = countResult
		//count = countResult
	}
	*/



    //Results
   	err := collection.Pipe(aggregate).AllowDiskUse().All(results)
	//iter := collection.Pipe(aggregate).AllowDiskUse().Iter()
	//for iter.Next(result){

	//}

	//err := iter.Close()
	if err != nil {
	    Debug(err)
		//panic("FindAggregate")

	}
	return err
}

func (db MongoDb) Find(model *Model, params Params, results interface{}) error {

	//Params
	mongoParams := db.formatParams(model, params)

	//DB
	collection, session := db.Collection(model, false)
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