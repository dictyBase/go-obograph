package manager

import (
	"context"
	"crypto/tls"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

// Session is a connected database client
type Session struct {
	client driver.Client
}

// NewSessionFromClient creates a new Session from an existing client
//   You could also do this
//      &Session{client}
//  Funny isn't it
func NewSessionFromClient(client driver.Client) *Session {
	return &Session{client}
}

// Connect is a constructor for new client
func Connect(host, user, password string, port int, istls bool) (*Session, error) {
	connConf := http.ConnectionConfig{
		Endpoints: []string{
			fmt.Sprintf("http://%s:%d", host, port),
		},
	}
	if istls {
		connConf.Endpoints = []string{
			fmt.Sprintf("https://%s:%d", host, port),
		}
		connConf.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	conn, err := http.NewConnection(connConf)
	if err != nil {
		return &Session{}, fmt.Errorf("could not connect %s", err)
	}
	client, err := driver.NewClient(
		driver.ClientConfig{
			Connection: conn,
			Authentication: driver.BasicAuthentication(
				user,
				password,
			),
		})
	if err != nil {
		return &Session{}, fmt.Errorf("could not get a client instance %s", err)
	}
	return &Session{client}, nil
}

// CurrentDB gets the default database(_system)
func (s *Session) CurrentDB() (*Database, error) {
	return s.getDatabase("_system")
}

// CreateDB creates database
func (s *Session) CreateDB(name string, opt *driver.CreateDatabaseOptions) error {
	ok, err := s.client.DatabaseExists(context.Background(), name)
	if err != nil {
		return fmt.Errorf("error in checking existence of database %s %s", name, err)
	}
	if !ok {
		_, err = s.client.CreateDatabase(context.Background(), name, opt)
		if err != nil {
			return fmt.Errorf("error in creating database %s %s", name, err)
		}
	}
	return nil
}

// DB gets the database
func (s *Session) DB(name string) (*Database, error) {
	return s.getDatabase(name)
}

// CreateUser creates user
func (s *Session) CreateUser(user, pass string) error {
	ok, err := s.client.UserExists(context.Background(), user)
	if err != nil {
		return fmt.Errorf("error in finding user %s", err)
	}
	if !ok {
		isActive := true
		_, err := s.client.CreateUser(context.Background(), user, &driver.UserOptions{Password: pass, Active: &isActive})
		if err != nil {
			return fmt.Errorf("error in creating user %s", err)
		}
	}
	return nil
}

// GrantDB grants user permission to a database
func (s *Session) GrantDB(database, user, grant string) error {
	ok, err := s.client.UserExists(context.Background(), user)
	if err != nil {
		return fmt.Errorf("error in finding user %s", err)
	}
	if !ok {
		return fmt.Errorf("user %s does not exist", user)
	}
	dbuser, err := s.client.User(context.Background(), user)
	if err != nil {
		return fmt.Errorf("error in getting user %s from database %s", user, err)
	}
	dbh, err := s.client.Database(context.Background(), database)
	if err != nil {
		return fmt.Errorf("cannot get a database instance %s", err)
	}
	err = dbuser.SetDatabaseAccess(context.Background(), dbh, getGrant(grant))
	if err != nil {
		return fmt.Errorf("error in setting database access %s", err)
	}
	return nil
}

func getGrant(g string) driver.Grant {
	var grnt driver.Grant
	switch g {
	case "rw":
		grnt = driver.GrantReadWrite
	case "ro":
		grnt = driver.GrantReadOnly
	default:
		grnt = driver.GrantNone
	}
	return grnt
}
func (s *Session) getDatabase(name string) (*Database, error) {
	ok, err := s.client.DatabaseExists(context.Background(), name)
	if err != nil {
		return &Database{}, err
	}
	if !ok {
		return &Database{}, fmt.Errorf("error in finding database %s", err)
	}
	dbh, err := s.client.Database(context.Background(), name)
	if err != nil {
		return &Database{}, fmt.Errorf("unable to get database instance %s", err)
	}
	return &Database{dbh}, nil
}
