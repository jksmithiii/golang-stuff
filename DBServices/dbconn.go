package DBServices

import (
	"sync"
	"database/sql"
	"log"
)

type TDBConnection struct {
	mu sync.RWMutex
	connectionString string
	driverName string
	db *sql.DB
	lasterror error
}

func (c *TDBConnection) Set(drivername string, conn string) {
	c.mu.Lock()
	{
		c.connectionString = conn
		c.driverName = drivername
	}
	c.mu.Unlock()
}

func (c *TDBConnection) SetLastError(err error) {
	c.mu.Lock()
	c.lasterror = err
	c.mu.Unlock()
}

func (c *TDBConnection) GetLastError() (err error) {
	c.mu.RLock()
	err = c.lasterror
	c.mu.RUnlock()
	return err
}

func (c *TDBConnection) GetDB() (db *sql.DB) {
	if c.db == nil {c.openDB()}
	db = c.db
	return db
}

func (c *TDBConnection) openDB() {
	if c.db == nil {
		c.mu.Lock()
		log.Println(c.driverName+":"+c.connectionString)
		c.db, c.lasterror = sql.Open(c.driverName,c.connectionString)
		c.mu.Unlock()
	}
}

func (c *TDBConnection) CloseDB() {
	if c.db == nil {
		c.mu.Lock()
		c.db.Close()
		c.mu.Unlock()
	}
}



