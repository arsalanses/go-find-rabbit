package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Mail struct {
	gorm.Model
	Mail string `json:"mail" gorm:"unique;not null"`
}

type CreateMailRequest struct {
	Mail string `json:"mail" binding:"required"`
}

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database")
	}

	err = database.AutoMigrate(&Mail{})

	if err != nil {
		return
	}

	DB = database
}

var CONNECTION *amqp.Connection
var CHANNEL *amqp.Channel

func ConnectQueue() {
	connection, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	CONNECTION = connection

	channel, err := connection.Channel()
	if err != nil {
		panic(err)
	}
	CHANNEL = channel

	queue, err := channel.QueueDeclare(
		"mail-sub", // name
		false,      // durable
		false,      // auto delete
		false,      // exclusive
		false,      // no wait
		nil,        // args
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(queue.Name)
}

func PostMail(c *gin.Context) {
	var input CreateMailRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mail := Mail{Mail: input.Mail}

	err := CHANNEL.Publish(
		"",         // exchange
		"mail-sub", // key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(mail.Mail),
		},
	)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusCreated, gin.H{"data": mail})
}

func main() {
	router := gin.Default()

	ConnectDatabase()
	ConnectQueue()
	defer CONNECTION.Close()
	defer CHANNEL.Close()

	router.POST("/api/v1/subscription", PostMail)

	err := router.Run("localhost:8080")

	if err != nil {
		return
	}
}
