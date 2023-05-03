package statuses

func StatusResponseFromStatus(status Status) StatusResponse {
	return StatusResponse{
		Content:  status.Content,
		Id:       status.Id,
		UserId:   status.UserId,
		MediaIds: &status.MediaIds,
	}
}
