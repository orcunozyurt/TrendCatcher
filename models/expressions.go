package models

import (
	"time"

	"github.com/rs/xid"
	"github.com/trendcatcher/database"
	"github.com/tuvistavie/structomap"
	"gopkg.in/mgo.v2/bson"
)

// DBTableTweetsPerMin collection name
const DBTableTweetsPerMin = "tweets_min"

// DBTableTweetsPerHour collection name
const DBTableTweetsPerHour = "tweets_hour"

// DBTableTweetsPerDay collection name
const DBTableTweetsPerDay = "tweets_day"

// DBTableTweetsPerMonth collection name
const DBTableTweetsPerMonth = "tweets_month"

// Expression structure
type Expression struct {
	ID        bson.ObjectId `json:"-" bson:"_id,omitempty"`
	URLToken  string        `json:"-" bson:"token,omitempty"`
	PostCount *int          `json:"post_count,omitempty" bson:"post_count,omitempty"`

	//Analysis  Analysis      `json:"analysis,omitempty" bson:"analysis,omitempty"`
	CreatedAt time.Time `json:"-" bson:"created_at,omitempty"`
	TweetedAt time.Time `json:"-" bson:"tweeted_at,omitempty"`
	UpdatedAt time.Time `json:"-" bson:"updated_at,omitempty"`
	DeletedAt time.Time `json:"-" bson:"deleted_at,omitempty"`
}

// Expressions array representation of Expression
type Expressions []Expression

// ListExpressions lists all expressions
func ListExpressions(query database.Query, paginationParams *database.PaginationParams) (*Expressions, error) {
	var result Expressions
	var dbtouse string

	if paginationParams == nil {
		paginationParams = database.NewPaginationParams()
		paginationParams.SortBy = "tweeted_at"
		dbtouse = DBTableTweetsPerMin
	} else if paginationParams.SortBy == "hour" {
		paginationParams.SortBy = "tweeted_at"
		dbtouse = DBTableTweetsPerHour
	} else if paginationParams.SortBy == "day" {
		paginationParams.SortBy = "tweeted_at"
		dbtouse = DBTableTweetsPerDay
	}

	err := database.Mongo.FindAll(dbtouse, query, &result, paginationParams)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetExpression an expression title with token
func GetExpression(query database.Query, period int) (*Expression, error) {
	var result Expression
	var dbtouse string

	switch period {
	case 0:
		dbtouse = DBTableTweetsPerMin

	case 1:
		dbtouse = DBTableTweetsPerHour

	case 2:
		dbtouse = DBTableTweetsPerDay

	default:
		dbtouse = DBTableTweetsPerMin
	}

	err := database.Mongo.FindOne(dbtouse, query, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create a new expression
func (expression *Expression) Create(period int) (*Expression, error) {

	var dbtouse string
	timenow := time.Now()
	var truncated time.Time

	switch period {
	case 0:
		dbtouse = DBTableTweetsPerMin
		truncated = timenow.Truncate(time.Minute)

	case 1:
		dbtouse = DBTableTweetsPerHour
		truncated = timenow.Truncate(time.Hour)

	case 2:
		dbtouse = DBTableTweetsPerDay
		truncated = time.Date(timenow.Year(), timenow.Month(), timenow.Day(), 0, 0, 0, 0, timenow.Location())

	default:
		dbtouse = DBTableTweetsPerMin
		truncated = timenow.Truncate(time.Minute)
	}

	expression.URLToken = xid.New().String()
	expression.CreatedAt = timenow
	expression.TweetedAt = truncated
	expression.UpdatedAt = expression.CreatedAt

	if err := database.Mongo.Insert(dbtouse, expression); err != nil {
		return nil, err
	}

	return expression, nil
}

// Update an expression
func (expression *Expression) Update(period int) (*Expression, error) {
	query := database.Query{}
	query["token"] = expression.URLToken
	var dbtouse string

	switch period {
	case 0:
		dbtouse = DBTableTweetsPerMin

	case 1:
		dbtouse = DBTableTweetsPerHour

	case 2:
		dbtouse = DBTableTweetsPerDay

	default:
		dbtouse = DBTableTweetsPerMin
	}

	expression.UpdatedAt = time.Now()

	change := database.DocumentChange{
		Update:    expression,
		ReturnNew: true,
	}

	result := &Expression{}
	err := database.Mongo.Update(dbtouse, query, change, result)

	return result, err
}

// Delete an expression
func (expression *Expression) Delete(period int) error {
	query := database.Query{}
	query["token"] = expression.URLToken
	var dbtouse string

	switch period {
	case 0:
		dbtouse = DBTableTweetsPerMin

	case 1:
		dbtouse = DBTableTweetsPerHour

	case 2:
		dbtouse = DBTableTweetsPerDay

	default:
		dbtouse = DBTableTweetsPerMin
	}

	expression.DeletedAt = time.Now()

	change := database.DocumentChange{
		Update:    expression,
		ReturnNew: true,
	}

	err := database.Mongo.Update(dbtouse, query, change, nil)

	return err
}

// ExpressionSerializer used in constructing maps to output JSON
type ExpressionSerializer struct {
	*structomap.Base
}

// NewExpressionSerializer creates a new ExpressionSerializer
func NewExpressionSerializer() *ExpressionSerializer {
	s := &ExpressionSerializer{structomap.New()}
	s.Pick("Postcount").
		PickFunc(func(t interface{}) interface{} {
			return t.(time.Time).Format(time.RFC3339)
		}, "TweetedAt", "UpdatedAt").
		AddFunc("ID", func(expression interface{}) interface{} {
			return expression.(Expression).URLToken
		})

	return s
}

// WithDeletedAt includes deletedAt field
func (s *ExpressionSerializer) WithDeletedAt() *ExpressionSerializer {
	s.PickFunc(func(t interface{}) interface{} {
		empty := time.Time{}
		if t.(time.Time) == empty {
			return nil
		}
		return t.(time.Time).Format(time.RFC3339)
	}, "DeletedAt")

	return s
}
