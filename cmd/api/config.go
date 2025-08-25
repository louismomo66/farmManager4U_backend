package main

import (
	"farm4u/data"
	"gorm.io/gorm"
	"log"
	"sync"
)

type Config struct {
	DB       *gorm.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	Wait     *sync.WaitGroup
	Models   data.Models

	ErrorChan     chan error
	ErrorChanDone chan bool
}
