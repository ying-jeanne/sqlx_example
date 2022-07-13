package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/profile"
)

// marshalize the row into struct, it could be less typed than standard sql library
// When use prepared statements, it could containes named parameter

const file string = "grafana.db"
const numrecords = 2000

type Team struct {
	ID        int
	Name      string         `db:"name"`
	OrgID     int            `db:"org_id"`
	CreatedAt time.Time      `db:"created"`
	UpdatedAt time.Time      `db:"updated"`
	Email     sql.NullString `db:"email"`
}

func insertTeam(db *sqlx.DB, team Team) error {
	// this is a transaction, a transaction should start with MustBegin, then end by commit
	// Inside of grafana code, instead of put session into the context, we can put the transaction into the context
	// tx := db.MustBegin()
	// tx.MustExec("INSERT INTO team (name, org_id, created, updated, email) VALUES ($1, $2, $3, $4, $5)", "wangyxxx", 0, time.Now(), time.Now(), "w.x@gmail.com")
	// named exec allows the user to insert into table with a struct object
	_, err := db.NamedExec("INSERT INTO team (name, org_id, created, updated, email) VALUES (:name, :org_id, :created, :updated, :email)", team)
	// tx.Commit()
	// here we would insert manually the fields instead of go struct, it is how sqlx is working today.
	// from the same project tho, there is a helper which helps us to build go struct https://github.com/jmoiron/modl/blob/master/modl_test.go
	// it could insert go struct instead of field by field
	return err
}

func getTeam(db *sqlx.DB, name string) (bool, Team, error) {
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
	found := true
	err := db.Get(&team1, "SELECT * FROM team WHERE Name=$1", name)
	if err == sql.ErrNoRows {
		found = false
		err = nil
	}
	return found, team1, err
	// for the field that could be nullable, we need to explicitely set it to type sql.NullString, otherwise we will get error when the field is not found
}

func updateTeam(db *sqlx.DB, team Team) error {
	_, err := db.NamedExec(`UPDATE team SET name=:name WHERE id =:id`, team)
	return err
}

func deleteTeam(db *sqlx.DB, name string) error {
	// the db.Rebind is to translate the ? to different presentation in different database type
	_, err := db.Exec("DELETE FROM team WHERE name=?", name)
	return err
	// if err != nil {
	// return 0, err
	// }
	// ids, err := rows.SliceScan()
	// if len(ids) > 0 {
	// 	return ids[0].(int), err
	// }
	// return 0, err
}

func main() {
	// create a sql.DB sqlite3 driver
	db, err := sqlx.Connect("sqlite3", file)
	if err != nil {
		log.Fatal(err)
	}

	// The test senarios:
	// 1. Create 1000 teams with go struct
	// 2. Verify after each creation so 1000 get with go struct casting
	// 3. Update 1000 teams with only name field
	// 4. Verify after update so 1000 get with go struct casting
	// 5. Delete 1000 teams with name field, return the id deleted
	// 6. Get 1000 teams which is nolonger existing

	// start timer
	start := time.Now()

	// CPU profiling by default related doc here: https://flaviocopes.com/golang-profiling/
	// defer profile.Start().Stop()

	// Memory profiling
	defer profile.Start(profile.MemProfile).Stop()

	for i := 0; i < numrecords; i++ {
		num := strconv.Itoa(i)
		team1 := Team{Name: "mynamee" + num, OrgID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		err = insertTeam(db, team1)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("the inserted team %+v\n", team1)
		_, team3, err := getTeam(db, team1.Name)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("the get team %+v\n", team3)
		team2 := Team{ID: team3.ID, OrgID: 0, Name: "princess" + num}
		err = updateTeam(db, team2)
		if err != nil {
			log.Fatal(err)
		}

		_, team4, err := getTeam(db, team2.Name)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("the updated team %+v\n", team4)

		err = deleteTeam(db, team4.Name)
		if err != nil {
			log.Fatal(err)
		}

		_, _, err = getTeam(db, team4.Name)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("the deleted team %+v\n", f)

	}
	fmt.Println("the result is: ", time.Since(start))
}
