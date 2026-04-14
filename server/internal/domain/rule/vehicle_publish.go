package rule

import (
	"vehivle/internal/domain/enum"
	"vehivle/internal/domain/model"
)

type PublishRequirements struct {
	HasCoverImage     bool
	CategoryEnabled   bool
	HasDetailImages   bool
	HasRequiredParams bool
}

func CanPublishVehicle(v *model.Vehicle, requirements *PublishRequirements) (ok bool, errs []string) {
	if v.Status != enum.VehicleStatusDraft {
		return false, []string{"仅草稿状态可发布"}
	}
	if v.Name == "" {
		errs = append(errs, "车型名称必填")
	}
	if v.CategoryID == nil || *v.CategoryID == "" {
		errs = append(errs, "一级分类必填")
	}
	if !requirements.CategoryEnabled {
		errs = append(errs, "分类未启用或不存在")
	}
	if !requirements.HasCoverImage {
		errs = append(errs, "封面图必填")
	}
	if !requirements.HasDetailImages {
		errs = append(errs, "详情图集必填")
	}
	if !requirements.HasRequiredParams {
		errs = append(errs, "参数模板未启用或必填参数项未填写完整")
	}
	return len(errs) == 0, errs
}
