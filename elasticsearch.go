package gomc

import (
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	"encoding/json"
	//"reflect"
	"strings"
	//"gopkg.in/mgo.v2"
    //"gopkg.in/mgo.v2/bson"
    //"time"
)

type Elasticsearch struct {}

type ElasticsearchParams struct {
	SortField string
	SortOrder bool
	From int
	Size int
	Query map[string]interface{}
	TermExact []map[string]interface{}
	Term []map[string]string
	Wildcard []map[string]string
}

func (db Elasticsearch) Connect(host string) (*elastic.Client, error) {

	client, err := elastic.NewClient(
		elastic.SetURL(host),
	)
	if err != nil {
	    // Handle error
	}

	return client, err
}

func (db Elasticsearch) formatParams(model *Model, params Params) (ElasticsearchParams){

	//Elasticsearch Params
	var elasticsearchParams ElasticsearchParams

	/*
	//Fields
	fields := make(map[string]int)
	for _, element := range params.Fields{
		fields[element] = 1
	}
	mongoDbParams.Fields = fields
	*/

	//Sort
	elasticsearchParams.SortField = model.Sort
	elasticsearchParams.SortOrder = true
	if params.Sort != "" {
		elasticsearchParams.SortField = params.Sort
	}

	
	//Page
	elasticsearchParams.From = 0
	elasticsearchParams.Size = model.Limit
	if params.Page > 0 {
		elasticsearchParams.From = ((params.Page - 1) * model.Limit)
		elasticsearchParams.Size = (params.Page * model.Limit)
	}

	
	//Query
	//elasticsearchParams.Query = params.Query
	for key,value := range params.Query{

		//Check For Match
		if strings.Contains(value.(string), "~"){
			termValue := strings.ToLower(strings.Replace(value.(string), "~", "", 1))
			term := map[string]string{
				key : termValue,
			}
			elasticsearchParams.Term = append(elasticsearchParams.Term, term)
		} else if strings.Contains(value.(string), "*"){
			wildcardValue := strings.ToLower(value.(string))
			wildcard := map[string]string{
				key : wildcardValue,
			}
			elasticsearchParams.Wildcard = append(elasticsearchParams.Wildcard, wildcard)
		} else {
			termExactKey := key + ".untouched"
			termExact := map[string]interface{}{
				termExactKey : value,
			}
			elasticsearchParams.TermExact = append(elasticsearchParams.TermExact, termExact)
		}
	}
	return elasticsearchParams
}

func (db Elasticsearch) Index(model *Model, result interface{}) error {
    
    // Create a client
	client, err := db.Connect(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Host + ":" + model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Port)
	
	_, err = client.Index().
    Index(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Database).
    Type(model.IndexDataUseTable).
    Id(model.IndexId).
    BodyJson(result).
    Do()
    
	if err != nil {
	    // Handle error
	    panic(err)
	}

    return err
} 

func (db Elasticsearch) Delete(model *Model, id string, result interface{}) error {
    
    // Create a client
	client, err := db.Connect(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Host + ":" + model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Port)
	
	_, err = elastic.NewDeleteService(client).
	Index(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Database).
	Type(model.IndexDataUseTable).
	Id(id).
	Do()
	fmt.Println(err)
	
    return err
} 

func (db Elasticsearch) Query(model *Model, params Params, results interface{}) error {
    
    // Create a client
	client, err := db.Connect(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Host + ":" + model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Port)
	if err != nil {
	    // Handle error
	    panic(err)
	}

	//Params
	elasticsearchParams := db.formatParams(model, params)
	
	//New Bool Query
	boolQuery := elastic.NewBoolQuery()

	//Term
	if len(elasticsearchParams.Term) > 0{
		for _, term := range elasticsearchParams.Term {
			for termKey, termValue := range term {
				termQuery := elastic.NewTermQuery(termKey, termValue)
				boolQuery.Must(termQuery)
			}
		}
	}

	//Wildcard
	if len(elasticsearchParams.Wildcard) > 0{
		for _, wildcard := range elasticsearchParams.Wildcard {
			for wildcardKey, wildcardValue := range wildcard {
				wildcardQuery := elastic.NewWildcardQuery(wildcardKey, wildcardValue)
				boolQuery.Must(wildcardQuery)
			}
		}
	}

	//Term Exact
	if len(elasticsearchParams.TermExact) > 0{
		for _, termExact := range elasticsearchParams.TermExact {
			for termExactKey, termExactValue := range termExact {
				termExactQuery := elastic.NewTermQuery(termExactKey, termExactValue)
				boolQuery.Filter(termExactQuery)
			}
		}
	}
	
	searchResult, err := client.Search().
    Index(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Database).   // search in index "twitter"
    Query(boolQuery).
    Sort(elasticsearchParams.SortField, elasticsearchParams.SortOrder).
    From(elasticsearchParams.From).Size(elasticsearchParams.Size).   // take documents 0-9
    Pretty(true).       // pretty print request and response JSON
    Do()
	if err != nil {
	    // Handle error
	    panic(err)
	}
	
	var newDataStringArray []string
	if searchResult.Hits != nil {
	    for _, hit := range searchResult.Hits.Hits {
			newDataStringArray = append(newDataStringArray,string(*hit.Source))
	    }
	}

	newDataString := "[" + strings.Join(newDataStringArray, ",") + "]"
	newDataByte := []byte(newDataString)
	
	err = json.Unmarshal(newDataByte, results)
	if err != nil {
        fmt.Println(err)
    }
    return err
} 
	