package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// loadUsers retrieves the Users struct from a file.
func loadUsers(filename string) (err error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading %s file: %s\n", filename, err.Error())
	} else {
		err = json.Unmarshal(b, &users)
		if err != nil {
			fmt.Printf("Error unmarshaling %s: %s\n", filename, err.Error())
		}
	}
	return
}

// saveUsers saves the Users struct to a file.
func saveUsers(filename string) (err error) {
	b, err := json.Marshal(users)
	if err != nil {
		fmt.Printf("Error marshaling %s: %s\n", filename, err.Error())
		return
	}
	err = ioutil.WriteFile(filename, b, 0644)
	if err != nil {
		fmt.Printf("Error writing %s file: %s\n", filename, err.Error())
		return
	}
	return
}

// loadSchedule saves the current schedule information to a file.
func loadSchedule(filename string) (err error) {
	var (
		b []byte
		data map[string]time.Time
	)

	if b, err = ioutil.ReadFile(filename); err != nil {
		fmt.Printf("Error reading %s file: %s\n", filename, err.Error())
		return
	}
	if err = json.Unmarshal(b, &data); err != nil {
		fmt.Printf("Error unmarshaling %s: %s\n", filename, err.Error())
		return
	}

	return
}

// saveSchedule saves the current data to a file in case we have to restart
func saveSchedule(schedule map[string]scheduleItem, filename string) (err error) {
	smap := make(map[string]time.Time)
	for k, v := range schedule {
		smap[k] = v.expireAt
	}

	b, err := json.Marshal(smap)
	if err != nil {
		fmt.Printf("Error marshaling %s: %s\n", filename, err.Error())
		return
	}
	err = ioutil.WriteFile(filename, b, 0644)
	if err != nil {
		fmt.Printf("Error writing %s file: %s\n", filename, err.Error())
		return
	}
	return
}

