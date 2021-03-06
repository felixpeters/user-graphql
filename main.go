package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
)

// User is a simple user object
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// UserList is list of users
var UserList []User
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func init() {
	user1 := User{ID: "a", Username: "Felix"}
	user2 := User{ID: "b", Username: "Jan"}
	user3 := User{ID: "c", Username: "Gregor"}
	UserList = append(UserList, user1, user2, user3)

	rand.Seed(time.Now().UnixNano())
}

// RandStringRunes creates random string
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"username": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var rootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type:        userType,
			Description: "Get single user",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				idQuery, isOK := params.Args["id"].(string)
				if isOK {
					for _, user := range UserList {
						if user.ID == idQuery {
							return user, nil
						}
					}
				}

				return User{}, nil
			},
		},
		"userList": &graphql.Field{
			Type:        graphql.NewList(userType),
			Description: "List of users",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return UserList, nil
			},
		},
	},
})

var rootMutation = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootMutation",
	Fields: graphql.Fields{
		"createUser": &graphql.Field{
			Type:        userType,
			Description: "Create new user",
			Args: graphql.FieldConfigArgument{
				"username": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				username, _ := params.Args["username"].(string)
				newID := RandStringRunes(8)
				newUser := User{
					ID:       newID,
					Username: username,
				}
				UserList = append(UserList, newUser)
				return newUser, nil
			},
		},
		"updateUser": &graphql.Field{
			Type:        userType,
			Description: "Update existing user",
			Args: graphql.FieldConfigArgument{
				"username": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				username, _ := params.Args["username"].(string)
				id, _ := params.Args["id"].(string)
				affectedUser := User{}

				for i := 0; i < len(UserList); i++ {
					if UserList[i].ID == id {
						UserList[i].Username = username
						affectedUser = UserList[i]
						break
					}
				}

				return affectedUser, nil
			},
		},
	},
})

var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    rootQuery,
	Mutation: rootMutation,
})

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v", result.Errors)
	}
	return result
}

type q struct {
	Query         string `json:"query"`
	OperationName string `json:"operationName"`
}

func setHTTPHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("HTTP headers set")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		next.ServeHTTP(w, r)
	})
}

func main() {
	gql := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			result := executeQuery(r.URL.Query()["query"][0], schema)
			json.NewEncoder(w).Encode(result)
		case "POST":
			w.Header().Set("Content-Type", "application/json")
			q := q{}
			body, _ := ioutil.ReadAll(r.Body)
			json.Unmarshal(body, &q)
			result := executeQuery(q.Query, schema)
			json.NewEncoder(w).Encode(result)
		case "OPTIONS":
		default:
			fmt.Fprintf(w, "Invalid request method")

		}
	}
	fmt.Println("Server is running on port 8080")
	fmt.Println("Get single user: curl -g 'http://localhost:8080/graphql?query={user(id:\"b\"){id,username}}'")
	fmt.Println("Create new user: curl -g 'http://localhost:8080/graphql?query=mutation+_{createUser(text:\"My+new+user\"){id,username}}'")
	fmt.Println("Update user: curl -g 'http://localhost:8080/graphql?query=mutation+_{updateUser(id:\"a\",username:\"Hans\"){id,username}}'")
	fmt.Println("Load user list: curl -g 'http://localhost:8080/graphql?query={userList{id,username}}'")

	graphQLHandler := http.HandlerFunc(gql)
	http.Handle("/graphql", setHTTPHeaders(graphQLHandler))
	http.ListenAndServe(":8080", nil)
}
