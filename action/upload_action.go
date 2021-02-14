package action

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/linklux/luxbox-server/component"
)

const CHUNK_SIZE = 1024

type UploadAction struct {
	component.ResourceCreator
}

func (this UploadAction) RequireUserAuth() bool { return true }

func (this UploadAction) New() IAction {
	return UploadAction{}
}

func (this UploadAction) Validate(request *Request) error {
	errs := make([]string, 0)

	// TODO Find a better way to validate params and their values
	if val, ok := request.Meta["resourceSize"]; !ok {
		errs = append(errs, "missing meta param 'resourceSize'")
	} else if resourceSize, ok := val.(float64); !ok { // 123 is a float64 when parsing json according to Go -_-
		errs = append(errs, "invalid value for 'resourceSize', must be numeric")
	} else if resourceSize < 1 {
		errs = append(errs, "invalid value for 'resourceSize', must be larger than 0")
	}

	if val, ok := request.Meta["resourceName"]; !ok {
		errs = append(errs, "missing or empty meta param: 'resourceName'")
	} else if resourceName, ok := val.(string); !ok {
		errs = append(errs, "invalid value for 'resourceName', must be string")
	} else {
		// If the optional overwrite meta param is included, use its value to
		// determine file overwrites otherwise, assume false
		overwrite := false
		if val, ok := request.Meta["overwrite"]; ok {
			if _, ok := val.(bool); !ok {
				errs = append(errs, "invalid value for 'overwrite', must be boolean")
			} else {
				overwrite = val.(bool)
			}
		}

		// If the file exists, and the overwrite flag is not given, fail validation
		if !overwrite && this.Exists("data", request.Meta["user"].(string), resourceName) {
			errs = append(errs, "resource exists, use the 'overwrite' flag to overwrite")
		}
	}

	if len(errs) > 0 {
		return errors.New(fmt.Sprintf("validation failed: %v", errs))
	}

	return nil
}

func (this UploadAction) Handle(request *Request) Response {
	f, err := this.Truncate("data", request.Meta["user"].(string), request.Meta["resourceName"].(string))
	if err != nil {
		return Response{-1, map[string]interface{}{
			"error": "failed to create resource",
		}}
	}

	resourceSize := uint64(request.Meta["resourceSize"].(float64))
	buf := make([]byte, 0, min(CHUNK_SIZE, resourceSize))

	r := bufio.NewReader(request.Conn)
	read, chunks := uint64(0), int(0)

	fmt.Printf("receiving data for resource '%s' (%d bytes) created by %s...",
		request.Meta["resourceName"].(string),
		resourceSize,
		request.Meta["user"].(string),
	)

	// Let the client know we're ready to start receiving data
	request.Conn.Write([]byte("ready\n"))

	var errMsg error = nil

	for read < resourceSize {
		n, err := r.Read(buf[:cap(buf)])
		if n < 1 {
			continue
		}

		chunks++
		read += uint64(n)

		if err != nil {
			errMsg = errors.New("error during buffer read: " + err.Error())
			break
		}

		if _, err := f.Write(buf[:n]); err != nil {
			errMsg = errors.New("error during file write: " + err.Error())
			break
		}

		// When reaching the last chunk, resize the buffer to match
		if resourceSize-read < CHUNK_SIZE && resourceSize-read > 0 {
			buf = make([]byte, 0, resourceSize-read)
		}
	}

	// When an error occured, delete the created resource and return an error
	if errMsg != nil {
		resource := f.Name()

		f.Close()
		os.Remove(resource)

		fmt.Printf(" failed\n")

		return Response{-1, map[string]interface{}{
			"error": fmt.Sprintf("resource creation failed, %s", errMsg.Error()),
		}}
	}

	f.Close()

	fmt.Printf(" done\n")

	return Response{3, map[string]interface{}{
		"message": "resource created",
	}}
}

func min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}
