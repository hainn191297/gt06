package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/zeromicro/go-zero/core/logx"
)

// ScyllaDBModel defines an interface for ScyllaDB operations
type ScyllaDBModel interface {
	Insert(ctx context.Context, tableName string, data map[string]interface{}) error
	Update(ctx context.Context, tableName string, filter map[string]interface{}, data map[string]interface{}) error
	Get(ctx context.Context, tableName string, filter map[string]interface{}) (map[string]interface{}, error)
	Delete(ctx context.Context, tableName string, filter map[string]interface{}) error
	Close() error
}

// scyllaDBModel is the implementation of ScyllaDBModel
type scyllaDBModel struct {
	session  *gocql.Session
	keyspace string
}

// NewScyllaDBModel initializes a new ScyllaDBModel instance
func NewScyllaDBModel(hosts []string, keyspace string, consistency gocql.Consistency) (ScyllaDBModel, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = consistency
	cluster.ConnectTimeout = 10 * time.Second
	cluster.Timeout = 5 * time.Second
	cluster.NumConns = 8
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	session, err := cluster.CreateSession()
	if err != nil {
		logx.Errorf("Failed to create ScyllaDB session: %v", err)
		return nil, err
	}

	logx.Info("Connected to ScyllaDB")
	return &scyllaDBModel{
		session:  session,
		keyspace: keyspace,
	}, nil
}

// Insert inserts a document into the specified table
func (s *scyllaDBModel) Insert(ctx context.Context, tableName string, data map[string]interface{}) error {
	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	placeholders := make([]string, 0, len(data))

	for k, v := range data {
		columns = append(columns, k)
		values = append(values, v)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ","),
		strings.Join(placeholders, ","),
	)

	return s.session.Query(query, values...).WithContext(ctx).Exec()
}

// Update updates documents in the specified table
func (s *scyllaDBModel) Update(ctx context.Context, tableName string, filter map[string]interface{}, data map[string]interface{}) error {
	setValues := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data)+len(filter))

	for k, v := range data {
		setValues = append(setValues, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	whereValues := make([]string, 0, len(filter))
	for k, v := range filter {
		whereValues = append(whereValues, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		tableName,
		strings.Join(setValues, ","),
		strings.Join(whereValues, " AND "),
	)

	return s.session.Query(query, args...).WithContext(ctx).Exec()
}

// Get retrieves a single document based on filter
func (s *scyllaDBModel) Get(ctx context.Context, tableName string, filter map[string]interface{}) (map[string]interface{}, error) {
	whereValues := make([]string, 0, len(filter))
	args := make([]interface{}, 0, len(filter))

	for k, v := range filter {
		whereValues = append(whereValues, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE %s LIMIT 1",
		tableName,
		strings.Join(whereValues, " AND "),
	)

	iter := s.session.Query(query, args...).WithContext(ctx).Iter()
	defer iter.Close()

	result := make(map[string]interface{})
	for iter.Scan(result) {
		return result, nil
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return nil, nil
}

// Delete removes documents from the specified table based on filter
func (s *scyllaDBModel) Delete(ctx context.Context, tableName string, filter map[string]interface{}) error {
	whereValues := make([]string, 0, len(filter))
	args := make([]interface{}, 0, len(filter))

	for k, v := range filter {
		whereValues = append(whereValues, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s",
		tableName,
		strings.Join(whereValues, " AND "),
	)

	return s.session.Query(query, args...).WithContext(ctx).Exec()
}

// Close closes the ScyllaDB session
func (s *scyllaDBModel) Close() error {
	if s.session != nil {
		s.session.Close()
	}
	return nil
}
