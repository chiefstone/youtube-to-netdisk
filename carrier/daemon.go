package carrier

import (
	"fmt"
	"github.com/cool2645/youtube-to-netdisk/model"
	"github.com/pkg/errors"
	"github.com/yanzay/log"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var queue = make(chan model.Task)
var kill = make(chan int64)
var runningId int64 = 0
var messages = make(chan model.Task)

func runDaemon() {
	go broadcast()
	go restore()
	for {
		task := <-queue
		runCarrier(task)
	}
}

func restore() {
	tasks, err := model.GetQueuingTasks(model.Db)
	if err != nil {
		log.Fatalf("failed while getting queuing tasks from db: %v", err)
		return
	}
	for _, task := range tasks {
		queue <- task
	}
}

func Push(task *model.Task) (ok bool, err error) {
	keywords, err := model.GetKeywords(model.Db)
	if err != nil {
		log.Fatalf("failed while getting keywords from db: %v", err)
		return
	}
	var hit = make([]string, 0)
	for _, v := range keywords {
		if strings.Contains(task.Title, v.Keyword) {
			hit = append(hit, v.Keyword)
		}
	}
	if len(hit) == 0 {
		task.Reason = "No keywords hit"
		task.State = "Rejected"
		err = model.CreateTask(model.Db, task)
		if err != nil {
			messages <- *task
		}
		ok = false
	} else {
		task.Reason = fmt.Sprintf("Keywords %v hit", hit)
		task.State = "Queuing"
		err = model.CreateTask(model.Db, task)
		if err != nil {
			go func() {
				queue <- *task
			}()
			messages <- *task
		}
		ok = true
	}
	return
}

func Cancel(task model.Task) (err error) {
	task.State = "Canceled"
	err = model.SaveTask(model.Db, &task)
	killTask(task.ID)
	return
}

func killTask(id int64) {
	go func() {
		kill <- id
	}()
}

func broadcast() {
	for {
		msg := <-messages
		for _, broadcaster := range broadcasters {
			broadcaster.Broadcast(msg)
		}
	}
}

func runCarrier(task model.Task) {
	runningId = task.ID
	task.State = "Downloading"
	model.SaveTask(model.Db, &task)

	// run command and read log
	task.State2 = runCmd(task.ID, config.TEMP_PATH, "python3", "-u", "lib/download.py", task.URL)
	var err error
	task.Log, err = ReadLog(task.ID)
	if err != nil {
		log.Error(err)
	}

	// get downloaded filename
	r := regexp.MustCompile(`fn:"(.*?)"`)
	p := r.FindStringSubmatch(task.Log)
	if len(p) >= 2 {
		task.FileName = p[1]
	}

	// got error or failed to read filename
	if task.State2 != "Finished" || len(task.FileName) == 0 {
		if task.State2 == "Killed" {
			// killed
			task.State = "Canceled"
		} else {
			task.State = "Exception"
		}
		model.SaveTask(model.Db, &task)
		messages <- task
		return
	}

	if len(uploaders) != 0 {
		task.State = "Uploading"
	} else {
		task.State = "Finished"
	}
	model.SaveTask(model.Db, &task)
	messages <- task

	ok := true
	if len(uploaders) != 0 {
		for _, uploader := range uploaders {
			uploaded, _ := uploader.Upload(task, messages)
			ok = ok && uploaded
		}
	}

	if ok {
		task.State = "Finished"
	} else {
		task.State = "Exception"
	}
	model.SaveTask(model.Db, &task)
	messages <- task
}

func ReadLog(taskID int64) (log string, err error) {
	fo := config.TEMP_PATH + "/" + strconv.FormatInt(taskID, 10) + "/" + strconv.FormatInt(taskID, 10) + ".log"
	b, err := ioutil.ReadFile(fo)
	if err != nil {
		err = errors.Wrapf(err, "Fail to read from file %s", fo)
		return
	}
	log += string(b)
	log += "\n"
	fe := config.TEMP_PATH + "/" + strconv.FormatInt(taskID, 10) + "/" + strconv.FormatInt(taskID, 10) + ".err.log"
	b, err = ioutil.ReadFile(fe)
	if err != nil {
		err = errors.Wrapf(err, "Fail to read from file %s", fe)
		return
	}
	log += string(b)
	return
}

func runCmd(id int64, tempPath string, c string, a ...string) (state string) {
	tempPath = tempPath + "/" + strconv.FormatInt(id, 10)
	os.MkdirAll(tempPath, os.ModePerm)
	fo, err := os.Create(tempPath + "/" + strconv.FormatInt(id, 10) + ".log")
	if err != nil {
		log.Error(err)
	}
	fe, err := os.Create(tempPath + "/" + strconv.FormatInt(id, 10) + ".err.log")
	if err != nil {
		log.Error(err)
	}
	cmd := exec.Command(c, a...)
	cmd.Dir = "static"
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	go io.Copy(fo, stdout)
	go io.Copy(fe, stderr)

	cmd.Start()
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	for {
		select {
		case id := <-kill:
			if id == runningId {
				if err := cmd.Process.Kill(); err != nil {
					log.Fatalf("failed to kill: %v", err)
					state = "Running"
				}
				log.Info("process killed")
				state = "Killed"
				return
			}
		case err := <-done:
			if err != nil {
				log.Errorf("process done with error = %v", err)
				state = "Error"
			} else {
				log.Info("process done gracefully without error")
				state = "Finished"
			}
			return
		}
	}
}
