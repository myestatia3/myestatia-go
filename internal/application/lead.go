// internal/application/lead_service.go
package application

import (
    "context"
    "bitbucket.org/statia/server/internal/domain/entity"

)

type LeadLogic struct {}


func (s LeadLogic) Create(ctx context.Context,name string, email string) (entity.Lead, error) {
    return entity.Lead{Name: name, Email: email}, nil
}