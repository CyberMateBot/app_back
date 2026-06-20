package ai

func init() {
	wavespeedThreeDModelCatalog = append(wavespeedThreeDModelCatalog, extendedThreeDModels...)
	for key, id := range extendedThreeDAliases {
		mediaModelAliases[key] = id
	}
}

var wavespeedThreeDModelCatalog []mediaModelDef

var extendedThreeDModels = []mediaModelDef{
	{
		ID: "tripo3d-v2.5-i2d", Label: "Tripo3D V2.5 Image", Group: "Tripo3D",
		Description: "3D-модель из одного изображения",
		TextSlug: "tripo3d/v2.5/image-to-3d",
		Provider: "wavespeed", Kind: "3d", RequiresImage: true,
	},
	{
		ID: "tripo3d-v2.5-multiview", Label: "Tripo3D V2.5 Multiview", Group: "Tripo3D",
		Description: "3D из 2–4 ракурсов",
		TextSlug: "tripo3d/v2.5/multiview-to-3d",
		Provider: "wavespeed", Kind: "3d", RequiresMultiImage: true,
	},
	{
		ID: "tripo3d-h3.1-t2d", Label: "Tripo3D H3.1 Text", Group: "Tripo3D",
		Description: "Генерация 3D по текстовому описанию",
		TextSlug: "tripo3d/h3.1/text-to-3d",
		Provider: "wavespeed", Kind: "3d",
	},
	{
		ID: "tripo3d-h3.1-i2d", Label: "Tripo3D H3.1 Image", Group: "Tripo3D",
		Description: "Детальная 3D-модель из фото",
		TextSlug: "tripo3d/h3.1/image-to-3d",
		Provider: "wavespeed", Kind: "3d", RequiresImage: true,
	},
	{
		ID: "hunyuan3d-v3-t2d", Label: "Hunyuan3D V3", Group: "Hunyuan3D",
		Description: "Tencent Hunyuan3D V3 text-to-3D",
		TextSlug: "wavespeed-ai/hunyuan3d-v3/text-to-3d",
		Provider: "wavespeed", Kind: "3d",
	},
	{
		ID: "hunyuan3d-v3.1-rapid", Label: "Hunyuan3D V3.1 Rapid", Group: "Hunyuan3D",
		Description: "Быстрая и экономичная 3D-генерация",
		TextSlug: "wavespeed-ai/hunyuan-3d-v3.1/text-to-3d-rapid",
		Provider: "wavespeed", Kind: "3d",
	},
	{
		ID: "meshy6-t2d", Label: "Meshy 6", Group: "Meshy",
		Description: "Высокое качество геометрии и текстур",
		TextSlug: "wavespeed-ai/meshy6/text-to-3d",
		Provider: "wavespeed", Kind: "3d",
	},
	{
		ID: "rodin-v2-i2d", Label: "Rodin V2", Group: "Hyper3D Rodin",
		Description: "Hyper3D Rodin V2 image-to-3D (10B)",
		TextSlug: "hyper3d/rodin-v2/image-to-3d",
		Provider: "wavespeed", Kind: "3d", RequiresImage: true,
	},
	{
		ID: "rodin-v2.5-i2d", Label: "Rodin V2.5", Group: "Hyper3D Rodin",
		Description: "Rodin V2.5 с расширенными настройками",
		TextSlug: "hyper3d/rodin-v2.5/image-to-3d",
		Provider: "wavespeed", Kind: "3d", RequiresImage: true,
	},
}

var extendedThreeDAliases = map[string]string{
	"tripo3d-v2.5-i2d":                             "tripo3d-v2.5-i2d",
	"tripo3d/v2.5/image-to-3d":                     "tripo3d-v2.5-i2d",
	"wavespeed-ai/tripo3d/v2.5/image-to-3d":          "tripo3d-v2.5-i2d",
	"tripo3d-v2.5-multiview":                         "tripo3d-v2.5-multiview",
	"tripo3d/v2.5/multiview-to-3d":                  "tripo3d-v2.5-multiview",
	"wavespeed-ai/tripo3d/v2.5/multiview-to-3d":    "tripo3d-v2.5-multiview",
	"tripo3d-h3.1-t2d":                             "tripo3d-h3.1-t2d",
	"tripo3d/h3.1/text-to-3d":                       "tripo3d-h3.1-t2d",
	"wavespeed-ai/tripo3d/h3.1/text-to-3d":         "tripo3d-h3.1-t2d",
	"tripo3d-h3.1-i2d":                             "tripo3d-h3.1-i2d",
	"tripo3d/h3.1/image-to-3d":                       "tripo3d-h3.1-i2d",
	"wavespeed-ai/tripo3d/h3.1/image-to-3d":        "tripo3d-h3.1-i2d",
	"hunyuan3d-v3-t2d":                         "hunyuan3d-v3-t2d",
	"wavespeed-ai/hunyuan3d-v3/text-to-3d":     "hunyuan3d-v3-t2d",
	"hunyuan3d-v3.1-rapid":                     "hunyuan3d-v3.1-rapid",
	"wavespeed-ai/hunyuan-3d-v3.1/text-to-3d-rapid": "hunyuan3d-v3.1-rapid",
	"meshy6-t2d":                               "meshy6-t2d",
	"wavespeed-ai/meshy6/text-to-3d":           "meshy6-t2d",
	"rodin-v2-i2d":                             "rodin-v2-i2d",
	"hyper3d/rodin-v2/image-to-3d":             "rodin-v2-i2d",
	"rodin-v2.5-i2d":                           "rodin-v2.5-i2d",
	"hyper3d/rodin-v2.5/image-to-3d":           "rodin-v2.5-i2d",
}
