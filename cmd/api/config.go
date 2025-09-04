package main

import (
	"farm4u/data"
	"log"
	"sync"

	"gorm.io/gorm"
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
