package pipeline

func FindHandlerEnds(allPipelines []*Pipeline) []*Pipeline {
	pipelines := make([]*Pipeline, 0, len(allPipelines))
	count := 0

	for _, pipeline := range allPipelines {
		if pipeline.End.IsHandler() {
			pipelines[count] = pipeline
			count++
		}
	}

	return pipelines
}

func FindServiceEnd(allPipelines []*Pipeline) *Pipeline {
	for _, pipeline := range allPipelines {
		if !pipeline.End.IsHandler() {
			return pipeline
		}
	}

	return nil
}
