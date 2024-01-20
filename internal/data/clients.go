package data

import "database/sql"

type Client struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func FindClientById(db *sql.DB, userId int) (Client, error) {
	var id int
	var name string
	err := db.QueryRow("SELECT * FROM clients where id = ?", userId).
		Scan(&id, &name)

	if err != nil {
		if err == sql.ErrNoRows {
			return Client{}, nil
		}
		return Client{}, err
	}
	cli := Client{
		Id:   id,
		Name: name,
	}
	return cli, nil
}
func AddClient(db *sql.DB, client Client) error {
	stmt, err := db.Prepare("INSERT INTO clients (id, name) values(?,?)")
	checkErr(err)

	_, err = stmt.Exec(client.Id, client.Name)
	if err != nil {
		return err
	}
	return nil
}
