package pgdocker

import (
	"context"
	"slices"
	"strconv"

	docker "github.com/gomatic/go-docker"
)

// Report is the inspected state of one managed port: its container (when
// one exists) and every managed data volume, newest first.
type Report struct {
	Instance     Instance
	Volumes      []docker.VolumeName
	Port         docker.Port
	HasContainer bool
}

// Inspect reports one port's managed container and volumes. A port with
// volumes but no container is a stopped-but-retained database; a port with
// neither yields an empty report.
func (m Manager) Inspect(ctx context.Context, port docker.Port) (Report, error) {
	if port == 0 {
		port = DefaultPort
	}
	reports, err := m.assemble(ctx, m.portFilter(port))
	if err != nil {
		return Report{}, err
	}
	for _, report := range reports {
		if report.Port == port {
			return report, nil
		}
	}
	return Report{Port: port}, nil
}

// List reports every managed port on the daemon, ordered by port.
func (m Manager) List(ctx context.Context) ([]Report, error) {
	return m.assemble(ctx, m.managedFilter())
}

// assemble joins labeled containers and volumes into per-port reports.
func (m Manager) assemble(ctx context.Context, filter docker.Labels) ([]Report, error) {
	containers, err := m.engine.Containers(ctx, docker.ContainerQuery{
		Labels:     filter,
		AllEnabled: true,
	})
	if err != nil {
		return nil, err
	}
	volumes, err := m.engine.Volumes(ctx, docker.VolumeQuery{Labels: filter})
	if err != nil {
		return nil, err
	}
	return m.mergeReports(containers, volumes), nil
}

// statusRunning is the Engine lifecycle state of a live container.
const statusRunning docker.ContainerStatus = "running"

// mergeReports groups containers and volumes by their port label. Snapshot
// volumes carry no port label, so they are filtered out here.
func (m Manager) mergeReports(
	containers []docker.ContainerSummary,
	volumes []docker.VolumeDetails,
) []Report {
	byPort := map[docker.Port]*Report{}
	for _, container := range containers {
		report := m.reportFor(byPort, m.portOf(container.Labels))
		report.HasContainer = true
		report.Instance = m.summaryInstance(container)
	}
	for _, volume := range volumes {
		if m.labelValue(volume.Labels, labelSnapshot) != "" {
			continue
		}
		report := m.reportFor(byPort, m.portOf(volume.Labels))
		report.Volumes = append(report.Volumes, volume.Name)
	}
	return sortedReports(byPort)
}

// reportFor finds or creates the report accumulating one port.
func (m Manager) reportFor(byPort map[docker.Port]*Report, port docker.Port) *Report {
	if existing, found := byPort[port]; found {
		return existing
	}
	created := &Report{Port: port}
	byPort[port] = created
	return created
}

// summaryInstance maps a container listing row onto the instance shape.
func (m Manager) summaryInstance(container docker.ContainerSummary) Instance {
	port := m.portOf(container.Labels)
	return Instance{
		Name:      container.Name,
		ID:        container.ID,
		Port:      port,
		Image:     container.Image,
		Status:    container.Status,
		Labels:    container.Labels,
		IsRunning: container.Status == statusRunning,
	}
}

// sortedReports flattens the port map ordered by port, with volume lists
// newest first (names embed a sortable timestamp).
func sortedReports(byPort map[docker.Port]*Report) []Report {
	ports := make([]int, 0, len(byPort))
	for port := range byPort {
		ports = append(ports, int(port))
	}
	slices.Sort(ports)
	reports := make([]Report, 0, len(ports))
	for _, port := range ports {
		report := byPort[docker.Port(port)]
		slices.Sort(report.Volumes)
		slices.Reverse(report.Volumes)
		reports = append(reports, *report)
	}
	return reports
}

// runningInstance finds the port's running managed container, failing with
// ErrNotRunning when none is up.
func (m Manager) runningInstance(ctx context.Context, port docker.Port) (Instance, error) {
	report, err := m.Inspect(ctx, port)
	if err != nil {
		return Instance{}, err
	}
	if !report.HasContainer || !report.Instance.IsRunning {
		return Instance{}, ErrNotRunning.With(nil, "port", strconv.Itoa(int(port)))
	}
	return report.Instance, nil
}
