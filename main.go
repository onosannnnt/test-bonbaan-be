package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
)

const (
	OmisePublicKey = "pkey_test_62o41s4lh9tra5ygw1t"
	OmiseSecretKey = "skey_test_62o41s51mezhmihco72"
)

func main() {
	app := fiber.New()

	app.Post("/webhook", handleWebhook)
	app.Post("/charge", createCharge)

	app.Listen(":3000")

}

type ChargeEvent struct {
	Data struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"data"`
}

func handleWebhook(c *fiber.Ctx) error {
	var event ChargeEvent

	if err := c.BodyParser(&event); err != nil {
		log.Println("Failed to parse webhook:", err)
		return c.Status(400).SendString("Invalid payload")
	}

	fmt.Println("Charge ID:", event.Data.ID)
	fmt.Println("Charge Status:", event.Data.Status)

	if event.Data.Status == "successful" {
		// Update order/payment status in database
		fmt.Println("Payment successful, update database!")
	} else if event.Data.Status == "failed" {
		fmt.Println("Payment failed, notify user!")
	}

	return c.SendStatus(200)
}

type ChargeRequest struct {
	Amount float64 `json:"amount"`
}

func createCharge(c *fiber.Ctx) error {
	var amount ChargeRequest
	if err := c.BodyParser(&amount); err != nil {
		return c.Status(400).SendString("Invalid payload")
	}
	if amount.Amount <= 20 {
		return c.Status(400).SendString("amount must be greater than or equal to ฿20 (2000 satangs)")
	}
	client, e := omise.NewClient(OmisePublicKey, OmiseSecretKey)
	if e != nil {
		log.Fatal(e)
	}
	source := &omise.Source{}
	err := client.Do(source, &operations.CreateSource{
		Amount:   int64(amount.Amount * 100),
		Currency: "thb",
		Type:     "promptpay",
	})
	if err != nil {
		return c.Status(400).SendString("Failed to create source")
	}

	charge := &omise.Charge{}
	err = client.Do(charge, &operations.CreateCharge{
		Amount:   source.Amount,
		Currency: source.Currency,
		Source:   source.ID,
	})
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":  err.Error(),
			"source": source,
		})
	}
	return c.JSON(charge)
}
