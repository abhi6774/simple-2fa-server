package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var (
	secretKey = "a21d94fdd217674b9232c60112d3142a"
	app       = fiber.New(fiber.Config{})
	secret    string
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: helloserver [options]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	greeting = flag.String("g", "Hello", "Greet with `greeting`")
	addr     = flag.String("addr", "localhost:8080", "address to serve")
)

func generateURI() (*otp.Key, error) {
	otpConfig := totp.GenerateOpts{
		Issuer:      "Glxy",
		AccountName: "abhi@mail.com",
		Secret:      []byte(secretKey),
		Period: 60,

	}

	otpURI, err := totp.Generate(otpConfig)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	fmt.Println("TOTP URI:", otpURI)
	secret = otpURI.Secret()

	return otpURI, nil
}

func getOTP() (string, error) {
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		fmt.Println("Error When generating code: ", err)
		return "", err
	}
	return code, nil
}


func main() {
	// Parse flags.
	flag.Usage = usage
	flag.Parse()

	// Parse and validate arguments (none).
	args := flag.Args()
	if len(args) != 0 {
		usage()
	}

	app.Get("/", greet)
	app.Get("/version", version)

	app.Get("/geturi", func(ctx *fiber.Ctx) error {
		otpURI, err := generateURI()
		if err != nil {
			log.Fatal(err)
			return err
		}
		ctx.Status(201).SendString(fmt.Sprintf("%v", otpURI))
		return nil
	})

	app.Get("/generateOTP", func(ctx *fiber.Ctx) error {
		code, err := getOTP()
		if err != nil {
			log.Fatal(err)
			return err
		}
		ctx.Status(201).JSON(map[string]any{
			"code": code,
		})
		return nil
	})

	app.Post("/validate", func(ctx *fiber.Ctx) error {
		type SecretPost struct {
			Code string `json:"code"`
		}

		body := ctx.Body()

		jsonData := SecretPost{}
		err := json.Unmarshal(body, &jsonData)
		fmt.Println("Check", secret, jsonData.Code)
		if err != nil {
			str := fmt.Sprintln("Error parsing the body to json: ", err)
			ctx.Status(500).SendString(str)
			return err
		}

		result := totp.Validate(jsonData.Code, secret)
		ctx.JSON(map[string]any{
			"verified": result,
		})
		return nil
	})

	log.Printf("serving http://%s\n", *addr)
	log.Fatal(app.Listen(*addr))
}

func version(ctx *fiber.Ctx) error {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		ctx.Status(500).SendString("no build information available")
		return nil
	}

	ctx.Context().SetContentType("text/html; charset=utf-8")
	fmt.Fprintf(ctx, "<!DOCTYPE html>\n<pre>\n")
	fmt.Fprintf(ctx, "%s\n", html.EscapeString(info.String()))
	return nil
}

func greet(ctx *fiber.Ctx) error {
	name := strings.Trim(ctx.BaseURL(), "/")
	if name == "" {
		name = "Gopher"
	}

	fmt.Fprintf(ctx, "<!DOCTYPE html>\n")
	fmt.Fprintf(ctx, "%s, %s!\n", *greeting, html.EscapeString(name))
	return nil
}
