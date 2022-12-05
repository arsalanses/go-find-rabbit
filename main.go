package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type Text struct {
	gorm.Model
	Text string `json:"text" gorm:"unique;not null"`
}

type CreateTextRequest struct {
	Text string `json:"text" binding:"required"`
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
		"text-sub", // name
		false,      // durable
		false,      // auto delete
		false,      // exclusive
		false,      // no wait
		nil,        // args
	)
	if err != nil {
		panic(err)
	}
	_ = queue
}

func PostText(c *gin.Context) {
	var input CreateTextRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	text := Text{Text: input.Text}

	err := CHANNEL.Publish(
		"",         // exchange
		"text-sub", // key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(text.Text),
		},
	)
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusCreated, gin.H{"data": text})
}

func main() {
	router := gin.Default()

	ConnectQueue()
	defer CONNECTION.Close()
	defer CHANNEL.Close()

	router.POST("/api/v1/subscription", PostText)

	err := router.Run("localhost:8080")

	if err != nil {
		return
	}
}
