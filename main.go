package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
)

var (
	OmisePublicKey string
	OmiseSecretKey string
)

func main() {
	_, filename, _, _ := runtime.Caller(0)
	currentFileDir := filepath.Dir(filename)

	// Walk up the directory tree to find the .env file
	var configPath string
	for {
		configPath = filepath.Join(currentFileDir, ".env")
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			break
		}
		parentDir := filepath.Dir(currentFileDir)
		if parentDir == currentFileDir {
			log.Fatalf(".env file not found")
			os.Exit(-1)
		}
		currentFileDir = parentDir
	}

	// Load the .env file
	err := godotenv.Load(configPath)
	if err != nil {
		log.Fatalf("Problem loading .env file: %v", err)
		os.Exit(-1)
	}
	OmisePublicKey = os.Getenv("OMISE_PUBLIC_KEY")
	OmiseSecretKey = os.Getenv("OMISE_SECRET_KEY")
	fmt.Println("OmisePublicKey:", OmisePublicKey)
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
