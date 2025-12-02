package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"simpleapi/internal/models"
	"simpleapi/internal/repository/sqlconnect"
	"strconv"
	"strings"
	"sync"
)

var teachers = make(map[int]models.Teacher)
var mutex = &sync.Mutex{}
var nextID = 1

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("METHOD :::: ", r.Method)
	switch r.Method {
	case http.MethodGet:
		// w.Write([]byte("Hello GET method on Teacher handler"))
		getTeachersHandler(w, r)
	case http.MethodPost:
		// w.Write([]byte("Hello POST method on Teacher handler"))
		addTeacherHandler(w, r)
	case http.MethodPut:
		w.Write([]byte("Hello PUT method on Teacher handler"))
	case http.MethodPatch:
		w.Write([]byte("Hello Patch method on Teacher handler"))
	case http.MethodDelete:
		w.Write([]byte("Hello DELETE method on Teacher handler"))
	}
}

func init() {
	teachers[nextID] = models.Teacher{
		ID:        nextID,
		FirstName: "John",
		LastName:  "Doe",
		Class:     "9A",
		Subject:   "Math",
	}
	nextID++
	teachers[nextID] = models.Teacher{
		ID:        nextID,
		FirstName: "Jane",
		LastName:  "Smith",
		Class:     "10A",
		Subject:   "Algebra",
	}
	nextID++
}

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")
	// fmt.Println(r.URL.Path)
	// fmt.Println(path)
	// fmt.Println(idStr)

	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			fmt.Println(err)
			return
		}
		teacher, exists := teachers[id]
		if !exists {
			http.Error(w, "Teacher not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(teacher)

	} else {

		firstName := r.URL.Query().Get("first_name")
		lastName := r.URL.Query().Get("last_name")

		teacherList := make([]models.Teacher, 0, len(teachers))
		for _, teacher := range teachers {
			if (firstName == "" || teacher.FirstName == firstName) && (lastName == "" || teacher.LastName == lastName) {
				teacherList = append(teacherList, teacher)
			}
		}
		response := struct {
			Status string           `json:"status"`
			Count  int              `json:"count"`
			Data   []models.Teacher `json:"data"`
		}{
			Status: "success",
			Count:  len(teachers),
			Data:   teacherList,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func addTeacherHandler(w http.ResponseWriter, r *http.Request) {
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
