package upload

import (
	"errors"
	"net/http"
)

type uploadError struct {
	code        int
	status      int
	description string
	err         error
}

func (e uploadError) Error() string {
	if e.err == nil {
		return e.description
	}
	return e.err.Error()
}

func (e uploadError) Unwrap() error {
	return e.err
}

func uploadErr(code int, status int, description string, err error) uploadError {
	if err == nil {
		err = errors.New(description)
	}
	return uploadError{code: code, status: status, description: description, err: err}
}

func formError(err error) uploadError {
	return uploadErr(101, http.StatusBadRequest, "invalid upload form", err)
}

func chartError(err error) uploadError {
	return uploadErr(102, http.StatusBadRequest, "invalid chart data", err)
}

func fileError(err error) uploadError {
	return uploadErr(301, http.StatusInternalServerError, "failed to save level files", err)
}

func fileTooBigError(err error) uploadError {
	return uploadErr(302, http.StatusBadRequest, "uploaded bgm file is too large", err)
}

func badBGMError(err error) uploadError {
	return uploadErr(303, http.StatusBadRequest, "invalid bgm format", err)
}

func bgmProcessError(err error) uploadError {
	return uploadErr(304, http.StatusInternalServerError, "failed to process bgm", err)
}

func badCoverError(err error) uploadError {
	return uploadErr(307, http.StatusBadRequest, "invalid cover format", err)
}

func coverProcessError(err error) uploadError {
	return uploadErr(308, http.StatusInternalServerError, "failed to process cover", err)
}

func chartConvertError(err error) uploadError {
	return uploadErr(305, http.StatusInternalServerError, "failed to convert chart", err)
}

func databaseError(err error) uploadError {
	return uploadErr(306, http.StatusInternalServerError, "failed to save database record", err)
}
