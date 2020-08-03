package v1

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	v1 "github.com/ryan0906/Memos/pkg/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	apiVersion = "v1"
)

type memoServiceServer struct {
	db *sql.DB
}

// NewMemoServiceServer Create Memo service
func NewMemoServiceServer(db *sql.DB) v1.MemoServiceServer {
	return &memoServiceServer{db: db}
}

func (sv *memoServiceServer) checkAPIVersion(api string) error {
	if len(api) > 0 {
		if api != apiVersion {
			return status.Errorf(codes.Unimplemented,
				"unsupported API version: request version %s cannot be resolved by API version %s", api, apiVersion)
		}
	}
	return nil
}

func (sv *memoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	conn, err := sv.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to connect to database, error: "+err.Error())
	}
	return conn, nil
}

// Create Memo
func (sv *memoServiceServer) Create(ctx context.Context, req *v1.CreateMemoRequest) (*v1.CreateMemoResponse, error) {
	if err := sv.checkAPIVersion(req.Api); err != nil {
		return nil, err
	}

	conn, err := sv.connect(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	reminder, err := ptypes.Timestamp(req.Memo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "reminder field has invalid format, error: "+err.Error())
	}

	res, err := conn.ExecContext(ctx,
		"INSERT INTO Memo(`Name`, `Description`, `Reminder`) VALUES(?, ?, ?)", req.Memo.Name, req.Memo.Description, reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to insert into Memo, error: "+err.Error())
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve inserted Memo, error: "+err.Error())
	}

	return &v1.CreateMemoResponse{
		Api: apiVersion,
		Id:  id,
	}, nil
}

func (sv *memoServiceServer) Read(ctx context.Context, req *v1.ReadMemoRequest) (*v1.ReadMemoResponse, error) {
	if err := sv.checkAPIVersion(req.Api); err != nil {
		return nil, err
	}

	conn, err := sv.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.QueryContext(ctx,
		"SELECT `ID`, `Name`, `Description`, `Reminder` FROM Memo WHERE `ID`=?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to select from Memo, error: "+err.Error())
	}

	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "Failed to retrieve data from Memo, error: "+err.Error())
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Memo with ID=%d is not found", req.Id))
	}

	var data v1.Memo
	var reminder time.Time

	if err := rows.Scan(&data.Id, &data.Name, &data.Description, &reminder); err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve fields from selected memo, error: "+err.Error())
	}

	data.Reminder, err = ptypes.TimestampProto(reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Invalid reminder type from selected Memo, error: "+err.Error())
	}

	if rows.Next() {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("Found multiple Memo with ID=%d", req.Id))
	}

	return &v1.ReadMemoResponse{
		Api:  apiVersion,
		Memo: &data,
	}, nil
}

func (sv *memoServiceServer) Update(ctx context.Context, req *v1.UpdateMemoRequest) (*v1.UpdateMemoResponse, error) {
	if err := sv.checkAPIVersion(req.Api); err != nil {
		return nil, err
	}

	conn, err := sv.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	reminder, err := ptypes.Timestamp(req.Memo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "reminder field has invalid format, error: "+err.Error())
	}

	res, err := conn.ExecContext(ctx,
		"UPDATE Memo SET `Name`=?, `Description`=?, `Reminder`=? WHERE `ID`=?",
		req.Memo.Name, req.Memo.Description, reminder, req.Memo.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to update Memo, error: "+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve affected Memo, error: "+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Memo with ID=%d is not found", req.Memo.Id))
	}

	return &v1.UpdateMemoResponse{
		Api:       apiVersion,
		IsUpdated: rows != 0,
	}, nil
}

func (sv *memoServiceServer) Delete(ctx context.Context, req *v1.DeleteMemoRequest) (*v1.DeleteMemoResponse, error) {
	if err := sv.checkAPIVersion(req.Api); err != nil {
		return nil, err
	}

	conn, err := sv.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res, err := conn.ExecContext(ctx, "DELETE FROM Memo WHERE `ID`=?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to delete from Memo, error: "+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve affected Memo, error: "+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Memo with ID=%d is not found", req.Id))
	}

	return &v1.DeleteMemoResponse{
		Api:       apiVersion,
		IsDeleted: rows != 0,
	}, nil
}

func (sv *memoServiceServer) ReadAll(ctx context.Context, req *v1.ReadAllRequest) (*v1.ReadAllResponse, error) {
	if err := sv.checkAPIVersion(req.Api); err != nil {
		return nil, err
	}

	conn, err := sv.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.QueryContext(ctx, "SELECT `ID`, `Name`, `Description`, `Reminder` FROM Memo")
	if err != nil {
		return nil, status.Error(codes.Unknown, "Failed to select from Memo, error: "+err.Error())
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "Failed to retrieve data from Memo, error: "+err.Error())
	}

	var reminder time.Time
	list := []*v1.Memo{}
	for rows.Next() {
		data := new(v1.Memo)
		if err := rows.Scan(&data.Id, &data.Name, &data.Description, &reminder); err != nil {
			return nil, status.Error(codes.Unknown, "Failed to retrieve fields from selected memo")
		}
		data.Reminder, err = ptypes.TimestampProto(reminder)
		if err != nil {
			return nil, status.Error(codes.Unknown, "reminder field has invalid format, error: "+err.Error())
		}
		list = append(list, data)
	}

	return &v1.ReadAllResponse{
		Api:   apiVersion,
		Memos: list,
	}, nil
}
