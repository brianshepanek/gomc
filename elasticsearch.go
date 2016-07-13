package gomc

import (
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	"encoding/json"
	"reflect"
	"strings"
	//"fmt"
	//"gopkg.in/mgo.v2"
    //"gopkg.in/mgo.v2/bson"
    //"time"
)

type Elasticsearch struct {}

type ElasticsearchNested struct {
	TermExact []map[string]interface{}
	TermsExact []map[string]interface{}
	Term []map[string]string
	Wildcard []map[string]string
	Prefix []map[string]string
}

type ElasticsearchParams struct {
	SortField string
	SortOrder bool
	Fields []string
	From int
	Size int
	Query map[string]interface{}
	TermExact []map[string]interface{}
	TermsExact []map[string]interface{}
	Term []map[string]string
	Wildcard []map[string]string
	Prefix []map[string]string
	GeoDistance map[string]float64
	Nested map[string]*ElasticsearchNested
}

func (db Elasticsearch) Connect(host string) (*elastic.Client, error) {

	client, err := elastic.NewClient(
		elastic.SetURL(host),
		elastic.SetSniff(false),
	)
	if err != nil {
	    panic(err)
	}

	return client, err
}

func (db Elasticsearch) formatParams(model *Model, params Params) (ElasticsearchParams){

	//Elasticsearch Params
	var elasticsearchParams ElasticsearchParams


	//Fields
	for _, element := range params.Fields{
		elasticsearchParams.Fields = append(elasticsearchParams.Fields, element)
	}

	//Sort
	elasticsearchParams.SortField = model.Sort
	elasticsearchParams.SortOrder = true
	if params.Sort != "" {
		if string(params.Sort[0]) == "-"{
			elasticsearchParams.SortField = strings.TrimPrefix(params.Sort, "-")
			elasticsearchParams.SortOrder = false
		} else {
			elasticsearchParams.SortField = params.Sort
		}
	}


	//Page
	elasticsearchParams.From = 0
	elasticsearchParams.Size = model.Limit
	if params.Page > 0 {
		elasticsearchParams.From = int((params.Page - 1) * model.Limit)
		elasticsearchParams.Size = int(params.Page * model.Limit)
	} else if params.Limit > 0 {
		elasticsearchParams.Size = int(params.Limit)
	}


	//Query
	//elasticsearchParams.Query = params.Query
	for key,value := range params.Query{
		if key != "near" && key != "radius" && key != "nested" {
			if reflect.TypeOf(value).String() == "string" {

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
				} else if strings.Contains(value.(string), "..."){
					prefixValue := strings.ToLower(strings.Replace(value.(string), "...", "", 1))
					prefix := map[string]string{
						key : prefixValue,
					}
					elasticsearchParams.Prefix = append(elasticsearchParams.Prefix, prefix)
				} else {
					termExactKey := key
					termExact := map[string]interface{}{
						termExactKey : value,
					}
					elasticsearchParams.TermExact = append(elasticsearchParams.TermExact, termExact)
				}
			} else {

				termsExactKey := key
				termsExact := map[string]interface{}{
					termsExactKey : value,
				}
				elasticsearchParams.TermsExact = append(elasticsearchParams.TermsExact, termsExact)
				/*
				valueArray := value.([]string)
				for _, arrayValue := range valueArray{
					termExact := map[string]interface{}{
						key : arrayValue,
					}
					elasticsearchParams.TermExact = append(elasticsearchParams.TermExact, termExact)
				}
				*/
			}
		}
	}
	//fmt.Println(elasticsearchParams)
	return elasticsearchParams
}

func (db Elasticsearch) IndexBulk(model *Model, docs []interface{}) error {

    // Create a client
	client, err := db.Connect(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Host + ":" + model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Port)
	bulkService := elastic.NewBulkService(client)
	
	for i := 0; i < len(docs); i++ {
		bulkIndexRequest := elastic.NewBulkIndexRequest().
		Index(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Database).
	    Type(model.IndexDataUseTable).
	    Id(model.IndexId).
		Doc(docs[i])

		bulkService.Add(bulkIndexRequest)
	}
	_, err = bulkService.Do()

    return err
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
	    fmt.Println(err)
	}

    return err
}

func (db Elasticsearch) CreateIndex(model *Model) error {

    // Create a client
	client, err := db.Connect(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Host + ":" + model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Port)

	_, err = elastic.NewIndicesCreateService(client).
	Index( model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Database).
	Do()
	if err != nil {
	    // Handle error
		fmt.Println("CreateIndex")
	    fmt.Println(err)
	}

    return err
}

func (db Elasticsearch) DeleteIndex(model *Model) error {

    // Create a client
	client, err := db.Connect(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Host + ":" + model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Port)

	var indices []string
	indices = append(indices, model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Database)

	//Check if Exists
	exists, err := elastic.NewIndicesExistsService(client).
	Index(indices).
	Do()
	if err != nil {
	    // Handle error
		fmt.Println("CheckIndex")
	    fmt.Println(err)
	}

	if exists {
		_, err = elastic.NewIndicesDeleteService(client).
		Index(indices).
		Do()
		if err != nil {
		    // Handle error
			fmt.Println("DeleteIndex")
		    fmt.Println(err)
		}
	}

    return err
}

func (db Elasticsearch) MapIndex(model *Model) error {

    // Create a client
	client, err := db.Connect(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Host + ":" + model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Port)

	_, err = elastic.NewIndicesPutMappingService(client).
	Index(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Database).
	Type(model.IndexDataUseTable).
	BodyJson(model.IndexMapping).
	Do()
	if err != nil {
	    // Handle error
		fmt.Println(err)
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
	if err != nil {
		fmt.Println(err)
	}
    return err
}

func (db Elasticsearch) Query(model *Model, params Params, results interface{}) error {

    // Create a client
	client, err := db.Connect(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Host + ":" + model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Port)
	if err != nil {
	    // Handle error
	    fmt.Println(err)
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

	//Prefix
	if len(elasticsearchParams.Prefix) > 0{
		for _, prefix := range elasticsearchParams.Prefix {
			for prefixKey, prefixValue := range prefix {
				prefixQuery := elastic.NewMatchPhrasePrefixQuery(prefixKey, prefixValue)
				boolQuery.Must(prefixQuery)
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

	//Terms Exact
	if len(elasticsearchParams.TermsExact) > 0{
		for _, termsExact := range elasticsearchParams.TermsExact {
			for termsExactKey, termsExactValue := range termsExact {
				termsExactQuery := elastic.NewTermsQuery(termsExactKey, termsExactValue)
				boolQuery.Filter(termsExactQuery)
			}
		}
	}

	//Near
	if _, ok := params.Query["near"]; ok {
		if _, ok := params.Query["radius"]; ok {
			cooridnatesValue := params.Query["near"].(map[string]float64)
			geoDistanceQuery := elastic.NewGeoDistanceQuery("loc.coordinates")
			geoDistanceQuery.Lat(cooridnatesValue["lat"])
			geoDistanceQuery.Lon(cooridnatesValue["lon"])
			geoDistanceQuery.Distance(params.Query["radius"].(string))
			boolQuery.Filter(geoDistanceQuery)
		}
	}


	//Nested
	if _, ok := params.Query["nested"]; ok {
		nestedQueries := params.Query["nested"].(map[string][]map[string]interface{})

		for rootKey, nestedArray := range nestedQueries {
			nestedBoolQuery := elastic.NewBoolQuery()
			for _, nestedQuery := range nestedArray{
				for key, value := range nestedQuery{
					if reflect.TypeOf(value).String() == "string" {

						//Check For Match
						if strings.Contains(value.(string), "~"){
							termValue := strings.ToLower(strings.Replace(value.(string), "~", "", 1))
							query := elastic.NewTermQuery(rootKey + "." + key, termValue)
							nestedBoolQuery.Must(query)
						} else if strings.Contains(value.(string), "*"){
							wildcardValue := strings.ToLower(value.(string))
							query := elastic.NewWildcardQuery(rootKey + "." + key, wildcardValue)
							nestedBoolQuery.Must(query)
						} else if strings.Contains(value.(string), "..."){
							prefixValue := strings.ToLower(strings.Replace(value.(string), "...", "", 1))
							query := elastic.NewMatchPhrasePrefixQuery(rootKey + "." + key, prefixValue)
							nestedBoolQuery.Must(query)
						} else {
							query := elastic.NewTermQuery(rootKey + "." + key, value)
							nestedBoolQuery.Filter(query)
						}
					} else {

						query := elastic.NewTermsQuery(rootKey + "." + key, value)
						nestedBoolQuery.Filter(query)
					}
				}
			}
			nestedQuery := elastic.NewNestedQuery(rootKey, nestedBoolQuery)
			boolQuery.Must(nestedQuery)
		}

	}

	//Count
	countResult, err := client.Count().
    Index(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Database).
    Query(boolQuery).
    Do()
	if err != nil {
	    // Handle error
	    fmt.Println(err)
	} else {
		model.Count = int(countResult)
		//count = countResult
	}

	//Search
	clientResourse := client.Search()
    clientResourse.Index(model.AppConfig.Databases[model.IndexDataUseDatabaseConfig].Database).
    Query(boolQuery).
    Sort(elasticsearchParams.SortField, elasticsearchParams.SortOrder).
    From(elasticsearchParams.From).
	Size(elasticsearchParams.Size).
    Pretty(true)

	searchResult, err := clientResourse.Do()

	if err != nil {
	    // Handle error
	    fmt.Println(err)
	}

	var newDataStringArray []string

	if searchResult != nil && searchResult.Hits != nil {
	    for _, hit := range searchResult.Hits.Hits {
			if hit.Fields != nil {

				jsonString, _ := json.Marshal(hit.Fields)
				newDataStringArray = append(newDataStringArray,string(jsonString))
			} else {
				newDataStringArray = append(newDataStringArray,string(*hit.Source))
			}
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
