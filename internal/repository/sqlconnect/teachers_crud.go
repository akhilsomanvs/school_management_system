package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"simpleapi/internal/models"
	"simpleapi/pkg/utils"
	"strconv"
	"strings"
)

func GetTeacherByID(id int) (models.Teacher, error) {
	db, err := ConnectDb()
	if err != nil {
		// http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()
	var teacher models.Teacher

	row := db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id)
	err = row.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
	if err == sql.ErrNoRows {
		// http.Error(w, "Teacher not found", http.StatusNotFound)
		return models.Teacher{}, utils.ErrorHandler(err, "Teacher not found")
	} else if err != nil {
		// http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Error fetching data")
	}
	return teacher, nil
}

func GetTeachersDbHandler(teachers []models.Teacher, r *http.Request) ([]models.Teacher, error) {

	db, err := ConnectDb()
	if err != nil {
		// http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
	var args []any

	query, args = addFilters(r, query)

	query = addSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println(err)
		// http.Error(w, "Database query error", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error retreiving data")
	}
	defer rows.Close()

	// teacherList := make([]models.Teacher, 0)

	for rows.Next() {
		var teacher models.Teacher
		err = rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			// http.Error(w, "Error fetching data", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "Error retreiving data")
		}
		teachers = append(teachers, teacher)
	}
	return teachers, nil
}

func AddTeachersDbHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDb()
	if err != nil {
		// http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()
	stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	if err != nil {
		fmt.Println(err)
		// http.Error(w, "Error in preparing SQL QUery", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error retreiving data")
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		if err != nil {
			fmt.Println(err)
			// http.Error(w, "Error inserting data into database", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "Error inserting data")
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			fmt.Println(err)
			// http.Error(w, "Error getting last inserted ID", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "Error inserting data")
		}
		newTeacher.ID = int(lastID)
		addedTeachers[i] = newTeacher
	}
	return addedTeachers, nil
}

func UpdateTeacher(id int, updatedTeacher models.Teacher) (models.Teacher, error) {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Error connecting to database")
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
			// http.Error(w, "Teacher not found", http.StatusNotFound)
			return models.Teacher{}, utils.ErrorHandler(err, "Error retreiving data")
		}
		// http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Error retreiving data")
	}

	updatedTeacher.ID = existingTeacher.ID
	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Class, updatedTeacher.Subject, updatedTeacher.ID)
	if err != nil {
		// http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Error updating teacher")
	}
	return updatedTeacher, nil
}

func PatchTeachers(updates []map[string]interface{}) error {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Unable to begin transaction", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Error retreiving data")
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			// http.Error(w, "Invalid or missing ID in update payload", http.StatusBadRequest)
			return utils.ErrorHandler(err, "Invalid ID")
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			tx.Rollback()
			log.Println(err)
			// http.Error(w, "Error converting Teacher ID", http.StatusBadRequest)
			return utils.ErrorHandler(err, "Error retreiving data")
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
				// http.Error(w, fmt.Sprintf("Teacher with ID %d not found", id), http.StatusNotFound)
				return utils.ErrorHandler(err, "Invalid ID")
			}
			// http.Error(w, "Error retrieving teacher", http.StatusInternalServerError)
			return utils.ErrorHandler(err, "Error updating data")
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
							return utils.ErrorHandler(err, "Error retreiving data")
						}
					}
					break
				}
			}
		}

		_, err = tx.Exec("UPDATE teachers SET first_name = ?, last_name = ?,email = ?,class = ?,subject = ? WHERE id = ?", teacherFromDb.FirstName, teacherFromDb.LastName, teacherFromDb.Email, teacherFromDb.Class, teacherFromDb.Subject, teacherFromDb.ID)
		if err != nil {
			tx.Rollback()
			// http.Error(w, "Error updating teacher", http.StatusInternalServerError)
			return utils.ErrorHandler(err, "Error retreiving data")
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Error retreiving data")
	}
	return nil
}

func PatchOneTeacher(id int, updates map[string]interface{}) (models.Teacher, error) {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Error connecting to database")
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
			// http.Error(w, "Teacher not found", http.StatusNotFound)
			return models.Teacher{}, utils.ErrorHandler(err, "Invalid ID")
		}
		// http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		//Apply udpates using Reflect
		return models.Teacher{}, utils.ErrorHandler(err, "Invalid ID")
	}

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
		// http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return models.Teacher{}, utils.ErrorHandler(err, "Error retreiving data")
	}
	return existingTeacher, nil
}

func DeleteOneTeacher(id int) error {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)

	if err != nil {
		log.Println(err)
		// http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Error retreiving data")
	}

	fmt.Println(result.RowsAffected())
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Error retrieving delete result", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Error retreiving data")
	}

	if rowsAffected == 0 {
		log.Println(err)
		// http.Error(w, "Teacher not found", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "Error retreiving data")
	}
	return nil
}

func DeleteTeachers(ids []int) ([]int, error) {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Error beginning transaction", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error retreiving data")
	}

	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id = ?")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		// http.Error(w, "Error preparing delete statement", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error retreiving data")
	}
	defer stmt.Close()

	deletedIds := []int{}

	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error retreiving data")
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println(err)
			tx.Rollback()
			// http.Error(w, "Error retrieving delete result", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "Error retreiving data")
		}
		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)
		}
		if rowsAffected < 1 {
			tx.Rollback()
			log.Printf("Teacher with ID %d not found", id)
			// http.Error(w, fmt.Sprintf("ID %d doesn not exist", id), http.StatusInternalServerError)
			continue
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		// http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error retreiving data")
	}

	if len(deletedIds) < 1 {
		// http.Error(w, "IDs do not exist", http.StatusNotFound)
		return nil, utils.ErrorHandler(err, "Error retreiving data")
	}
	return deletedIds, nil
}

//------------------------------

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
