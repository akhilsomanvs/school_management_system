package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"simpleapi/internal/models"
	"simpleapi/internal/repository/sqlconnect"
	"strconv"
	"strings"
)

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("METHOD :::: ", r.Method)
	switch r.Method {
	case http.MethodGet:
		// w.Write([]byte("Hello GET method on Teacher handler"))
		GetTeachersHandler(w, r)
	case http.MethodPost:
		// w.Write([]byte("Hello POST method on Teacher handler"))
		AddTeacherHandler(w, r)
	case http.MethodPut:
		// w.Write([]byte("Hello PUT method on Teacher handler"))
		UpdateTeacherHandler(w, r)
	case http.MethodPatch:
		//w.Write([]byte("Hello Patch method on Teacher handler"))
		PatchTeachersHandler(w, r)
	case http.MethodDelete:
		// w.Write([]byte("Hello DELETE method on Teacher handler"))
		DeleteOneTeacherHandler(w, r)
	}
}

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	}
	return validFields[field]
}

func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
	var args []any

	query, args = addFilters(r, query)

	query = addSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	teacherList := make([]models.Teacher, 0)

	for rows.Next() {
		var teacher models.Teacher
		err = rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			http.Error(w, "Error fetching data", http.StatusInternalServerError)
			return
		}
		teacherList = append(teacherList, teacher)
	}
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(teacherList),
		Data:   teacherList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetOneTeacherHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	idStr := r.PathValue("id ")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	var teacher models.Teacher
	row := db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id)
	err = row.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)

}

func addSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]

	if len(sortParams) > 0 {
		query += " ORDER BY"
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidSortField(field) || !isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				query += ","
			}
			query += " " + field + " " + order
		}
	}
	return query
}

func addFilters(r *http.Request, query string) (string, []interface{}) {
	var args []any
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"class":      "class",
		"subject":    "subject",
	}

	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + dbField + " = ?"
			args = append(args, value)
		}
	}
	return query, args
}

func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var newTeachers []models.Teacher

	err = json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error in preparing SQL QUery", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error inserting data into database", http.StatusInternalServerError)
			return
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error getting last inserted ID", http.StatusInternalServerError)
			return
		}
		newTeacher.ID = int(lastID)
		addedTeachers[i] = newTeacher
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}

	json.NewEncoder(w).Encode(response)

}

// PUT /teacher/{id}
func UpdateTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		return
	}
	var updatedTeacher models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(
		&existingTeacher.ID,
		&existingTeacher.FirstName,
		&existingTeacher.LastName,
		&existingTeacher.Email,
		&existingTeacher.Class,
		&existingTeacher.Subject,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Teacher not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return
	}

	updatedTeacher.ID = existingTeacher.ID
	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Class, updatedTeacher.Subject, updatedTeacher.ID)
	if err != nil {
		http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacher)
}

// PATCH /teachers/{id}
func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var updates []map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to begin transaction", http.StatusInternalServerError)
		return
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			http.Error(w, "Invalid or missing ID in update payload", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			tx.Rollback()
			log.Println(err)
			http.Error(w, "Error converting Teacher ID", http.StatusBadRequest)
			return
		}

		var teacherFromDb models.Teacher
		err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(
			&teacherFromDb.ID,
			&teacherFromDb.FirstName,
			&teacherFromDb.LastName,
			&teacherFromDb.Email,
			&teacherFromDb.Class,
			&teacherFromDb.Subject,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				http.Error(w, fmt.Sprintf("Teacher with ID %d not found", id), http.StatusNotFound)
				return
			}
			http.Error(w, "Error retrieving teacher", http.StatusInternalServerError)
			return
		}

		//Apply updates using Reflect
		teacherVal := reflect.ValueOf(&teacherFromDb).Elem()
		teacherType := teacherVal.Type()

		for k, v := range update {
			if k == "id" {
				continue //skip updating the id field
			}
			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherType.Field(i)
				if field.Tag.Get("json") == k+",omitempty" {
					fieldVal := teacherVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(v)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							log.Printf("Cannot convert %v to %v", val.Type(), fieldVal.Type())
							return
						}
					}
					break
				}
			}
		}

		_, err = tx.Exec("UPDATE teachers SET first_name = ?, last_name = ?,email = ?,class = ?,subject = ? WHERE id = ?", teacherFromDb.FirstName, teacherFromDb.LastName, teacherFromDb.Email, teacherFromDb.Class, teacherFromDb.Subject, teacherFromDb.ID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Error updating teacher", http.StatusInternalServerError)
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

func PatchOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).Scan(
		&existingTeacher.ID,
		&existingTeacher.FirstName,
		&existingTeacher.LastName,
		&existingTeacher.Email,
		&existingTeacher.Class,
		&existingTeacher.Subject,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Teacher not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return
	}

	//Apply udpates using Reflect
	teacherVal := reflect.ValueOf(&existingTeacher).Elem()
	teacherType := teacherVal.Type()

	for k, v := range updates {
		for i := 0; i < teacherVal.NumField(); i++ {
			field := teacherType.Field(i)
			fieldTag := field.Tag.Get("json")
			if strings.Contains(fieldTag, ",") {
				fieldTag = strings.Split(fieldTag, ",")[0]
			}
			if fieldTag == k {
				if teacherVal.Field(i).CanSet() {
					fieldVal := teacherVal.Field(i)
					fieldVal.Set(reflect.ValueOf(v).Convert(fieldVal.Type()))
				}
			}
		}
	}

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		existingTeacher.FirstName,
		existingTeacher.LastName,
		existingTeacher.Email,
		existingTeacher.Class,
		existingTeacher.Subject,
		existingTeacher.ID)
	if err != nil {
		http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTeacher)
}

// DELETE TEacher
func DeleteOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)

	if err != nil {
		log.Println(err)
		http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
		return
	}

	fmt.Println(result.RowsAffected())
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving delete result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Println(err)
		http.Error(w, "Teacher not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "Teacher successfully deleted",
		ID:     id,
	}

	json.NewEncoder(w).Encode(response)

}

func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var ids []int
	err = json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error beginning transaction", http.StatusInternalServerError)
		return
	}

	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id = ?")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		http.Error(w, "Error preparing delete statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	deletedIds := []int{}

	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println(err)
			tx.Rollback()
			http.Error(w, "Error retrieving delete result", http.StatusInternalServerError)
			return
		}
		if rowsAffected == 0 {
			log.Printf("Teacher with ID %d not found", id)
			continue
		}
		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	if len(deletedIds) < 1 {
		http.Error(w, "IDs do not exist", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status     string `json:"status"`
		DeletedIDs []int  `json:"deleted_ids"`
	}{
		Status:     "Teachers successfully deleted",
		DeletedIDs: deletedIds,
	}

	json.NewEncoder(w).Encode(response)
}
