package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

// HistoryResponse maps to the Fyers API v3 historical data structure
type HistoryResponse struct {
	Status  string      `json:"s"`
	Candles [][]float64 `json:"candles"`
	Message string      `json:"message"`
}

func main() {
	appID := "YOUR_APP_ID"
	accessToken := "YOUR_ACCESS_TOKEN"
	symbol := "NSE:RELIANCE-EQ" 

	client := resty.New()

	var result HistoryResponse
	resp, err := client.R().
		SetHeader("Authorization", fmt.Sprintf("%s:%s", appID, accessToken)).
		SetQueryParams(map[string]string{
			"symbol":      symbol,
			"resolution":  "D", // "D" for Daily, ideal for swing analysis
			"date_format": "1", // 1 enables standard yyyy-mm-dd format
			"range_from":  "2023-01-01",
			"range_to":    "2023-12-31",
			"cont_flag":   "1", // 1 maintains continuous data for adjusted splits/dividends
		}).
		SetResult(&result).
		Get("https://api-t1.fyers.in/data/history")

	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	if resp.IsError() || result.Status != "ok" {
		log.Fatalf("API Error: %s, Message: %s", resp.Status(), result.Message)
	}

	fmt.Printf("Retrieved %d daily candles for %s.\n", len(result.Candles), symbol)
	
	// Print the first 5 records
	for i, c := range result.Candles {
		if i >= 5 {
			break
		}
		// The API returns the timestamp in epoch seconds as the first array element
		date := time.Unix(int64(c[0]), 0).Format("2006-01-02")
		fmt.Printf("Date: %s | Open: %.2f | High: %.2f | Low: %.2f | Close: %.2f | Vol: %.0f\n",
			date, c[1], c[2], c[3], c[4], c[5])
	}
}




// updated
package main

import (
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
)

// TokenResponse maps the JSON returned when exchanging the auth code
type TokenResponse struct {
	Status      string `json:"s"`
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
}

func main() {
	// Initialize a new Fiber app
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// This route matches the "Redirect URI" you set in your broker's dashboard
	// Example: http://127.0.0.1:3000/callback
	app.Get("/callback", func(c *fiber.Ctx) error {
		// 1. Extract the auth code sent by the broker after successful login
		authCode := c.Query("auth_code")
		if authCode == "" {
			return c.Status(400).SendString("Error: No auth_code found in the URL")
		}

		fmt.Printf("Received Auth Code from broker: %s\n", authCode)

		// 2. Exchange this code for the actual Access Token
		token, err := getAccessToken(authCode)
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Token exchange failed: %v", err))
		}

		// 3. Success! In a real system, you would save this token to PostgreSQL or a local file
		fmt.Printf("SUCCESS! Access Token Acquired: %s\n", token)

		return c.SendString("Authentication complete! You can close this browser tab and return to your terminal.")
	})

	fmt.Println("Waiting for broker redirect... Please log in via your browser.")
	fmt.Println("Listening on http://127.0.0.1:3000")
	
	// Start the server
	log.Fatal(app.Listen(":3000"))
}

// getAccessToken uses Resty to swap the auth code for a usable API token
func getAccessToken(authCode string) (string, error) {
	client := resty.New()
	var result TokenResponse

	// Note: Check your specific broker's documentation for the exact payload required here.
	// For Fyers v3, this typically requires an appIdHash (SHA-256 of appId:appSecret).
	appIDHash := "YOUR_GENERATED_APP_ID_HASH" 

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"grant_type": "authorization_code",
			"appIdHash":  appIDHash,
			"code":       authCode,
		}).
		SetResult(&result).
		Post("https://api-t1.fyers.in/api/v3/validate-authcode")

	if err != nil {
		return "", err
	}

	if resp.IsError() || result.Status != "ok" {
		return "", fmt.Errorf("API Error: %s, Message: %s", resp.Status(), result.Message)
	}

	return result.AccessToken, nil
}



package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
)

// openBrowser launches the default system browser to the specified URL
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin": // macOS
		cmd = "open"
	default: // Linux variants
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}



// updated
func main() {
	appID := "YOUR_APP_ID"
	redirectURI := "http://127.0.0.1:3000/callback"
	
	// Construct the Fyers OAuth login URL
	// Note: Replace state and other parameters as required by Fyers API v3 docs
	authURL := fmt.Sprintf(
		"https://api-t1.fyers.in/api/v3/generate-authcode?client_id=%s&redirect_uri=%s&response_type=code&state=sample_state",
		appID, redirectURI,
	)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Get("/callback", func(c *fiber.Ctx) error {
		authCode := c.Query("auth_code")
		if authCode == "" {
			return c.Status(400).SendString("Error: No auth_code found")
		}

		fmt.Printf("Received Auth Code: %s\n", authCode)
		// token, err := getAccessToken(authCode) ... (from previous example)
		
		return c.SendString("Authentication complete! You can close this tab.")
	})

	// Use a goroutine to wait a split second for the server to start before opening the browser
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Opening browser for authentication...")
		err := openBrowser(authURL)
		if err != nil {
			fmt.Printf("Could not auto-open browser. Please manually navigate to: \n%s\n", authURL)
		}
	}()

	fmt.Println("Listening on http://127.0.0.1:3000")
	log.Fatal(app.Listen(":3000"))
}




package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BrokerCredentials represents the table where we store API tokens
type BrokerCredentials struct {
	BrokerName  string    `gorm:"primaryKey"` // e.g., "FYERS"
	AccessToken string    `gorm:"not null"`
	UpdatedAt   time.Time // Helps your trading engine know if the token is fresh for today
}

// connectDB initializes the PostgreSQL connection using GORM
func connectDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=supersecret dbname=swing_trading port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&BrokerCredentials{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// saveToken securely stores or updates the access token
func saveToken(db *gorm.DB, token string) error {
	cred := BrokerCredentials{
		BrokerName:  "FYERS",
		AccessToken: token,
		UpdatedAt:   time.Now(),
	}

	// We use an Upsert (Update if exists, Insert if it does not).
	// Since Fyers tokens expire daily, we just overwrite yesterday's token.
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "broker_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"access_token", "updated_at"}),
	}).Create(&cred)

	return result.Error
}

func main() {
	// 1. Initialize DB connection
	db := connectDB()
	fmt.Println("Successfully connected to PostgreSQL via Podman.")

	// Simulated Token acquired from your Fiber OAuth callback
	simulatedToken := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..." 

	// 2. Save the token to the database
	err := saveToken(db, simulatedToken)
	if err != nil {
		log.Fatalf("Could not save token: %v", err)
	}

	fmt.Println("Token successfully saved to the database. Your trading engine is ready to go.")
	
	// 3. Verify it was saved by retrieving it
	var activeCred BrokerCredentials
	db.First(&activeCred, "broker_name = ?", "FYERS")
	fmt.Printf("Retrieved Token from DB (Last Updated: %s)\n", activeCred.UpdatedAt.Format("2006-01-02 15:04:05"))
}



