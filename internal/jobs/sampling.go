package jobs

func sampling(producer any) bool {
	return producer == producerShadow || producer == producerJudge
}
