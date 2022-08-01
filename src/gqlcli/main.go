package main

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hasura/go-graphql-client"
)

type Profile struct {
	Name graphql.String
	Id   graphql.ID
}

type Profiles struct {
	Profiles []Profile
}

type ClientWrapper struct {
	Client *graphql.Client
}

func (c *ClientWrapper) Open(url string) {
	c.Client = graphql.NewClient(url, nil)
}

func (c *ClientWrapper) AddRecord2(name string, id int, table string) (string, error) {
	args := makeArgs(name, id)
	query := "mutation{insert_" + table + "_one(object: {" + strings.Join(args, ", ") + "}){name, id}}"
	fmt.Println(query)
	var m struct {
		AddItem Profile
	}
	mi := prepareExecArg(&m, `graphql:"insert_`+table+`_one"`)
	err := c.Client.Exec(context.Background(), query, mi, nil)
	return fmt.Sprint(m.AddItem.Id), err
}

func (c *ClientWrapper) AddRecord(name string, id int) (string, error) {
	return c.AddRecord2(name, id, "profiles")
}

func (c *ClientWrapper) RemoveByName2(name string, table string) ([]graphql.ID, error) {
	fncName := "delete_" + table
	var m struct {
		Res struct {
			Returning []Profile
		}
	}
	query := "mutation{" + fncName + "(where: {name: {_eq: \"" + name + "\"}}){returning {id}}}"
	fmt.Println(query)
	mi := prepareExecArg(&m, `graphql:"`+fncName+`"`)
	err := c.Client.Exec(context.Background(), query, mi, nil)
	var ids []graphql.ID
	for _, i := range m.Res.Returning {
		ids = append(ids, i.Id)
	}
	return ids, err
}

func (c *ClientWrapper) RemoveByName(name string) ([]graphql.ID, error) {
	return c.RemoveByName2(name, "profiles")
}

func (c *ClientWrapper) QueryAll2(table string) ([]Profile, error) {
	var m Profiles
	arg := prepareExecArg(&m, `graphql:"`+table+`"`)
	str, err := graphql.ConstructQuery(arg, nil)
	fmt.Println("query=" + str)
	err = c.Client.Query(context.Background(), arg, nil)
	return m.Profiles, err
}

func (c *ClientWrapper) QueryAll() ([]Profile, error) {
	return c.QueryAll2("profiles")
}

func main() {
	var c ClientWrapper
	c.Open("http://localhost:8080/v1/graphql")

	queryAll := func() {
		pr, err := c.QueryAll()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("query all=" + fmt.Sprint(pr))
	}

	queryAll()

	var q struct {
		Profiles []Profile `graphql:"profiles(where: {id: {_eq: 1}})"`
	}
	str, err := graphql.ConstructQuery(&q, nil)
	fmt.Println("query=" + str)
	err = c.Client.Query(context.Background(), &q, nil)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println("response=" + fmt.Sprint(q))

	var m struct {
		AddFedya struct {
			Name graphql.String
			Id   graphql.ID
		} `graphql:"insert_profiles_one(object: {name: \"fedya\"})"`
	}
	str, err = graphql.ConstructQuery(&m, nil)
	fmt.Println("mutation=" + str)
	err = c.Client.Mutate(context.Background(), &m, nil)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println("response=" + fmt.Sprint(m))

	queryAll()

	var md struct {
		DelFedya struct {
			Returning []Profile
		} `graphql:"delete_profiles(where: {name: {_eq: \"fedya\"}})"`
	}
	str, err = graphql.ConstructQuery(&md, nil)
	fmt.Println("mutation=" + str)
	err = c.Client.Mutate(context.Background(), &md, nil)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println("response=" + fmt.Sprint(md))

	queryAll()

	str, err = c.AddRecord("Mityai", -1)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("addRecord returned %s\n", str)

	queryAll()

	str, err = c.AddRecord2("Mityai", -1, "profiles")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("addRecord2 returned %s\n", str)
	queryAll()

	ids, err := c.RemoveByName2("Mityai", "profiles")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("removeByName returned: %s\n", fmt.Sprint(ids))

	queryAll()
}

func prepareExecArg[D any](data *D, tag string) interface{} {
	value := reflect.ValueOf(data).Elem()
	t := value.Type()
	sf := make([]reflect.StructField, 0)
	sf = append(sf, t.Field(0))
	sf[0].Tag = reflect.StructTag(tag)
	newType := reflect.PointerTo(reflect.StructOf(sf))
	newValue := reflect.ValueOf(data).Convert(newType)
	i := newValue.Interface()
	return i
}

func makeArgs(name string, id int) []string {
	args := []string{
		"name: \"" + name + "\"",
	}
	if id > 0 {
		args = append(args, "id: "+fmt.Sprint(id))
	}
	return args
}
