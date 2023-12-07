package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-redis/redis"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/sirupsen/logrus"
)

// Post represents the data model
type Post struct {
	PostID      string `json:"PostID"`
	PostUser    string `json:"PostUser"`
	PostMessage string `json:"PostMessage"`
}

var postType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Post",
		Fields: graphql.Fields{
			"PostID":      &graphql.Field{Type: graphql.String},
			"PostUser":    &graphql.Field{Type: graphql.String},
			"PostMessage": &graphql.Field{Type: graphql.String},
		},
	},
)

var rootQuery = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"post": &graphql.Field{
				Type: postType,
				Args: graphql.FieldConfigArgument{
					"postID": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					postID, ok := p.Args["postID"].(string)
					if !ok {
						return nil, nil
					}

					// Retrieve data from Redis
					result, err := redisClient.Get(postID).Result()
					if err != nil {
						logrus.Errorf("Error fetching data from Redis: %s", err)
						return nil, err
					}

					// Unmarshal the JSON string from Redis into a Post object
					var post Post
					err = json.Unmarshal([]byte(result), &post)
					if err != nil {
						logrus.Errorf("Error unmarshalling JSON: %s", err)
						return nil, err
					}

					return post, nil
				},
			},
		},
	},
)

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: rootQuery,
	},
)

var redisClient *redis.Client

func init() {
	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: "redis-service:6379",
	})
}

func main() {
	// Set up logging
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stdout)

	// Example log usage
	logrus.Info("Server starting...")

	// Set up GraphQL handler
	h := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})

	// Define GraphQL endpoint
	http.Handle("/graphql", h)

	// Start the server
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}
