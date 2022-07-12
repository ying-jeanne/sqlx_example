package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// marshalize the row into struct, it could be less typed than standard sql library
// When use prepared statements, it could containes named parameter

const file string = "grafana.db"

type Team struct {
	ID        int
	Name      string    `db:"name"`
	OrgID     int       `db:"org_id"`
	CreatedAt time.Time `db:"created"`
	UpdatedAt time.Time `db:"updated"`
	Email     sql.NullString
}

func setTeam(db *sqlx.DB, team Team) error {
	// this is a transaction, a transaction should start with MustBegin, then end by commit
	// Inside of grafana code, instead of put session into the context, we can put the transaction into the context
	tx := db.MustBegin()
	// tx.MustExec("INSERT INTO team (name, org_id, created, updated, email) VALUES ($1, $2, $3, $4, $5)", "wangyxxx", 0, time.Now(), time.Now(), "w.x@gmail.com")
	// named exec allows the user to insert into table with a struct object
	_, err := tx.NamedExec("INSERT INTO team (name, org_id, created, updated) VALUES (:name, :org_id, :created, :updated)", team)
	if err != nil {
		log.Fatal(err)
		tx.Rollback()
	}
	tx.Commit()
	return err
}

func getTeam(db *sqlx.DB, name string) (Team, error) {
	// teams := []Team{}
	// err := db.Select(&teams, "SELECT * FROM team ORDER BY name ASC")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// here to cast the type into go struct directly we still need the tag for the fields that are not having the same name
	// but the error is threw correctly when it could not found the corresponding field, so it is still better than xorm?
	// for _, team := range teams {
	// 	fmt.Printf("%#v\n", team)
	// }

	// get one single result
	team1 := Team{}
	err := db.Get(&team1, "SELECT * FROM team WHERE Name=$1", name)
	return team1, err
	// for the field that could be nullable, we need to explicitely set it to type sql.NullString, otherwise we will get error when the field is not found
}

func deleteTeam(db *sqlx.DB, name string) error {
	// the db.Rebind is to translate the ? to different presentation in different database type
	res, err := db.Exec(db.Rebind("DELETE FROM team WHERE NAME=?"), name)
	row, _ := res.RowsAffected()
	fmt.Printf("%#v\n", row)
	return err
}

func updateTeam(db *sqlx.DB, team Team) error {
	_, err := db.NamedExec(`UPDATE team SET name=:name, org_id=:org_id WHERE id =:id`, team)
	return err
}

func main() {
	db, err := sqlx.Connect("sqlite3", file)
	if err != nil {
		log.Fatal(err)
	}
	team1 := Team{Name: "myname7", OrgID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_ = setTeam(db, team1)

	team3, err := getTeam(db, team1.Name)
	if err != nil {
		log.Fatal(err)
	}

	team2 := Team{ID: team3.ID, OrgID: 0, Name: "princess"}
	err = updateTeam(db, team2)
	if err != nil {
		log.Fatal(err)
	}
	team4, err := getTeam(db, team2.Name)
	if err != nil {
		log.Fatal(err)
	}
	err = deleteTeam(db, team4.Name)
	if err != nil {
		log.Fatal(err)
	}
}
