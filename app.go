package phly

// Support for phly-based applications

func RunPipeline() (PipelineResult, error) {
	p, err := LoadPipeline(`C:\work\dev\go\src\github.com\hackborn\phly\phly\data\scale.json`)
	if err != nil {
		return PipelineResult{}, err
	}
	args := RunArgs{}
	return p.Run(args)
}
