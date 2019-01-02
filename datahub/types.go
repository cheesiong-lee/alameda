package datahub

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
)

type podMetricExtended metric.PodMetric

func (p podMetricExtended) datahubPodMetric() datahub_v1alpha1.PodMetric {

	var (
		datahubPodMetric datahub_v1alpha1.PodMetric
	)

	datahubPodMetric = datahub_v1alpha1.PodMetric{
		NamespacedName: &datahub_v1alpha1.NamespacedName{
			Namespace: string(p.Namespace),
			Name:      string(p.PodName),
		},
	}

	for _, containerMetric := range p.ContainersMetricMap {
		containerMetricExtended := containerMetricExtended(containerMetric)
		datahubContainerMetric := containerMetricExtended.datahubContainerMetric()
		datahubPodMetric.ContainerMetrics = append(datahubPodMetric.ContainerMetrics, &datahubContainerMetric)
	}

	return datahubPodMetric
}

type containerMetricExtended metric.ContainerMetric

func (c containerMetricExtended) NumberOfDatahubMetricDataNeededProducing() int {
	return 2
}

func (c containerMetricExtended) datahubContainerMetric() datahub_v1alpha1.ContainerMetric {

	var (
		metricDataChan = make(chan datahub_v1alpha1.MetricData)

		datahubContainerMetric datahub_v1alpha1.ContainerMetric
	)

	datahubContainerMetric = datahub_v1alpha1.ContainerMetric{
		Name: string(c.ContainerName),
	}

	go c.produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_CONTAINER_CPU_USAGE_SECONDS_PERCENTAGE, c.CPUMetircs, metricDataChan)
	go c.produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_CONTAINER_MEMORY_USAGE_BYTES, c.MemoryMetrics, metricDataChan)

	for i := 0; i < c.NumberOfDatahubMetricDataNeededProducing(); i++ {
		receivedMetricData := <-metricDataChan
		datahubContainerMetric.MetricData = append(datahubContainerMetric.MetricData, &receivedMetricData)
	}

	return datahubContainerMetric
}

func (c containerMetricExtended) produceDatahubMetricDataFromSamples(metricType datahub_v1alpha1.MetricType, samples []metric.Sample, metricDataChan chan<- datahub_v1alpha1.MetricData) {

	var (
		datahubMetricData datahub_v1alpha1.MetricData
	)

	datahubMetricData = datahub_v1alpha1.MetricData{
		MetricType: metricType,
	}

	for _, sample := range samples {

		// TODO: Send error to caller
		googleTimestamp, err := ptypes.TimestampProto(sample.Timestamp)
		if err != nil {
			googleTimestamp = nil
		}

		datahubSample := datahub_v1alpha1.Sample{Time: googleTimestamp, NumValue: sample.Value}
		datahubMetricData.Data = append(datahubMetricData.Data, &datahubSample)
	}

	metricDataChan <- datahubMetricData
}