package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

// untuk men trigger ke variable Conn di function DatabaseConnect sehingga varr Conn bisa di panggil di main.go
var Conn *pgx.Conn

func DatabaseConnect() {
	// koneksi databae
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	databaseUrl := "postgres://postgres:siburian@localhost:5000/personal-web"

	var err error

	Conn, err = pgx.Connect(context.Background(), databaseUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to database:%v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database connected")

}
