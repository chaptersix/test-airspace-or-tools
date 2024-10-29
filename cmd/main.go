package main

import (
	"fmt"
	"log"
	"time"

	"github.com/airspacetechnologies/or-tools/go/ortools/constraintsolver"
	"github.com/golang/protobuf/ptypes/duration"
)

type DataModelCPDPTW struct {
	timeMatrix        [][]int64
	timeWindows       [][]int64
	pickupDeliveries  [][]int
	allowedVehicles   [][]int
	vehicleCapacities []int64
	demands           []int
	depot             int
}

var (
	timeMatrix = [][]int64{
		{0, 6, 9, 8, 7, 3, 6, 2, 3, 2, 6, 6, 4, 4, 5, 9, 7},
		{6, 0, 8, 3, 2, 6, 8, 4, 8, 8, 13, 7, 5, 8, 12, 10, 14},
		{9, 8, 0, 11, 10, 6, 3, 9, 5, 8, 4, 15, 14, 13, 9, 18, 9},
		{8, 3, 11, 0, 1, 7, 10, 6, 10, 10, 14, 6, 7, 9, 14, 6, 16},
		{7, 2, 10, 1, 0, 6, 9, 4, 8, 9, 13, 4, 6, 8, 12, 8, 14},
		{3, 6, 6, 7, 6, 0, 2, 3, 2, 2, 7, 9, 7, 7, 6, 12, 8},
		{6, 8, 3, 10, 9, 2, 0, 6, 2, 5, 4, 12, 10, 10, 6, 15, 5},
		{2, 4, 9, 6, 4, 3, 6, 0, 4, 4, 8, 5, 4, 3, 7, 8, 10},
		{3, 8, 5, 10, 8, 2, 2, 4, 0, 3, 4, 9, 8, 7, 3, 13, 6},
		{2, 8, 8, 10, 9, 2, 5, 4, 3, 0, 4, 6, 5, 4, 3, 9, 5},
		{6, 13, 4, 14, 13, 7, 4, 8, 4, 4, 0, 10, 9, 8, 4, 13, 4},
		{6, 7, 15, 6, 4, 9, 12, 5, 9, 6, 10, 0, 1, 3, 7, 3, 10},
		{4, 5, 14, 7, 6, 7, 10, 4, 8, 5, 9, 1, 0, 2, 6, 4, 8},
		{4, 8, 13, 9, 8, 7, 10, 3, 7, 4, 8, 3, 2, 0, 4, 5, 6},
		{5, 12, 9, 14, 12, 6, 6, 7, 3, 3, 4, 7, 6, 4, 0, 9, 2},
		{9, 10, 18, 6, 8, 12, 15, 8, 13, 9, 13, 3, 4, 5, 9, 0, 9},
		{7, 14, 9, 16, 14, 8, 5, 10, 6, 5, 4, 10, 8, 6, 2, 9, 0},
	}

	timeWindows = [][]int64{
		{0, 5},   // depot
		{7, 12},  // 1
		{10, 15}, // 2
		{16, 18}, // 3
		{10, 13}, // 4
		{0, 5},   // 5
		{5, 10},  // 6
		{0, 4},   // 7
		{5, 10},  // 8
		{0, 3},   // 9
		{10, 16}, // 10
		{10, 15}, // 11
		{0, 5},   // 12
		{5, 10},  // 13
		{7, 8},   // 14
		{10, 15}, // 15
		{11, 15}, // 16
	}
)

func main() {
	n := time.Now()
	routing, data, manager, cleanup := setup()
	defer cleanup()

	// Setting first solution heuristic.
	searchParameters := constraintsolver.DefaultRoutingSearchParameters()
	searchParameters.TimeLimit = &duration.Duration{Seconds: 5}
	searchParameters.SolutionLimit = 100
	searchParameters.FirstSolutionStrategy = constraintsolver.FirstSolutionStrategy_PATH_CHEAPEST_ARC
	searchParameters.LocalSearchMetaheuristic = constraintsolver.LocalSearchMetaheuristic_GUIDED_LOCAL_SEARCH

	// Asynchronously cancel search.
	// go func() {
	// 	time.Sleep(2 * time.Millisecond)
	// 	routing.CancelSearch()
	// }()

	// Solve the problem.
	solution := routing.SolveWithParameters(searchParameters)

	// Print solution on console.
	printSolution(data, manager, routing, &solution)
	fmt.Println(time.Since(n))
}

// Print the solution.
// data: Data of the problem.
// manager: Index manager used.
// routing: Routing solver used.
func printSolution(data DataModelCPDPTW, manager constraintsolver.RoutingIndexManager, routing constraintsolver.RoutingModel, solution *constraintsolver.Assignment) {
	log.Printf("Solver status: %v", routing.GetStatus())

	// Display dropped nodes
	dropped := "Dropped nodes:"
	for n := range routing.Size() {
		if routing.IsStart(n) || routing.IsEnd(n) {
			continue
		}
		if (solution != nil && (*solution).Value(routing.NextVar(n)) == n) ||
			(solution == nil && routing.NextVar(n).Value() == n) {
			dropped += fmt.Sprintf(" %v", manager.IndexToNode(n))
		}
	}
	log.Printf(dropped)
	log.Println()

	// Display routes
	timeDimension := routing.GetDimensionOrDie("Time")
	capacityDimension := routing.GetDimensionOrDie("Capacity")
	var totalTime, totalLoad, tMin, tMax, sMin, sMax, capacity int64
	for vehicleId := 0; vehicleId < len(data.vehicleCapacities); vehicleId++ {
		index := routing.Start(vehicleId)
		log.Printf("Route for vehicle %v:", vehicleId)
		var route string
		for !routing.IsEnd(index) {
			timeVar := timeDimension.CumulVar(index)
			slackVar := timeDimension.SlackVar(index)
			capacityVar := capacityDimension.CumulVar(index)
			if solution == nil {
				tMin = timeVar.Min()
				tMax = timeVar.Max()
				sMin = slackVar.Min()
				sMax = slackVar.Max()
				capacity = capacityVar.Value()
			} else {
				tMin = (*solution).Min(timeVar)
				tMax = (*solution).Max(timeVar)
				sMin = (*solution).Min(slackVar)
				sMax = (*solution).Max(slackVar)
				capacity = (*solution).Value(capacityVar)
			}
			route += fmt.Sprintf("%v Time(%v, %v) Slack(%v, %v) Load:%v -> ",
				manager.IndexToNode(index),
				tMin,
				tMax,
				sMin,
				sMax,
				capacity)
			if solution == nil {
				index = routing.NextVar(index).Value()
			} else {
				index = (*solution).Value(routing.NextVar(index))
			}
		}
		timeVar := timeDimension.CumulVar(index)
		capacityVar := capacityDimension.CumulVar(index)
		if solution == nil {
			tMin = timeVar.Min()
			tMax = timeVar.Max()
			capacity = capacityVar.Value()
		} else {
			tMin = (*solution).Min(timeVar)
			tMax = (*solution).Max(timeVar)
			capacity = (*solution).Value(capacityVar)
		}
		log.Printf("%v%v Time(%v, %v) Load:%v\n",
			route,
			manager.IndexToNode(index),
			tMin,
			tMax,
			capacity)
		log.Printf("Time of the route: %vmin\n", tMin)
		log.Printf("Load of the route: %v\n", capacity)
		log.Println()
		totalTime += tMin
		totalLoad += capacity
	}
	log.Printf("Total time of all routes: %vmin\n", totalTime)
	log.Printf("Total load of all routes: %v\n", totalLoad)
	log.Printf("Advanced usage:")
	log.Printf("Problem solved in %vms\n", routing.Solver().WallTime())
}

func setup() (constraintsolver.RoutingModel, DataModelCPDPTW, constraintsolver.RoutingIndexManager, func()) {

	// Solvable pickup deliveries
	solvablePickupDeliveries := [][]int{
		{6, 2},
		{9, 14},
		{5, 10},
	}

	// Allow all vehicles for each pickup/delivery
	allowedVehicles := [][]int{
		{0, 1, 2, 3}, // depot
		{},           // 1
		{0, 1, 2, 3}, // 2
		{},           // 3
		{},           // 4
		{0, 1, 2, 3}, // 5
		{0, 1, 2, 3}, // 6
		{},           // 7
		{},           // 8
		{0, 1, 2, 3}, // 9
		{0, 1, 2, 3}, // 10
		{},           // 11
		{},           // 12
		{},           // 13
		{0, 1, 2, 3}, // 14
		{},           // 15
		{},           // 16
	}

	// Instantiate the data problem.
	data := DataModelCPDPTW{
		timeMatrix:        timeMatrix,
		timeWindows:       timeWindows,
		pickupDeliveries:  solvablePickupDeliveries,
		allowedVehicles:   allowedVehicles,
		vehicleCapacities: []int64{15, 15, 15, 15},
		demands:           []int{0, 0, -1, 0, 0, 1, 1, 0, 0, 1, -1, 0, 0, 0, -1, 0, 0},
		depot:             0,
	}

	// Create Routing Index Manager
	starts := []int{0, 0, 0, 0}
	ends := []int{0, 0, 0, 0}
	manager := constraintsolver.NewRoutingIndexManager(len(data.timeMatrix), len(data.vehicleCapacities),
		starts, ends)

	// Create Routing Model.
	routing := constraintsolver.NewRoutingModel(manager, constraintsolver.DefaultRoutingModelParameters())

	// Allow dropping stops
	penalty := int64(1000)
	for i := 1; i < len(data.timeMatrix); i++ {
		routing.AddDisjunction([]int64{manager.NodeToIndex(i)}, penalty)
	}

	// Create and register a time transit callback.
	f := func(fromIndex, toIndex int64) int64 {
		// Convert from routing variable Index to time matrix NodeIndex.
		fromNode := manager.IndexToNode(fromIndex)
		toNode := manager.IndexToNode(toIndex)
		return data.timeMatrix[fromNode][toNode]
	}
	w := constraintsolver.NewGoRoutingTransitCallback2Wrapper(f)
	transitCallbackIndex := routing.RegisterTransitCallback(w.Wrap())

	// Define cost of each arc.
	routing.SetArcCostEvaluatorOfAllVehicles(transitCallbackIndex)

	// Add Time constraint.
	routing.AddDimension(transitCallbackIndex, // transit callback index
		int64(30),  // allow waiting time
		int64(360), // maximum time per vehicle
		false,      // Don't force start cumul to zero
		"Time")
	timeDimension := routing.GetDimensionOrDie("Time")

	// Create and register a capacity transit callback.
	f2 := func(fromIndex int64) int64 {
		// Convert from routing variable Index to time matrix NodeIndex.
		fromNode := manager.IndexToNode(fromIndex)
		return int64(data.demands[fromNode])
	}
	w2 := constraintsolver.NewGoRoutingTransitCallback1Wrapper(f2)
	capacityCallbackIndex := routing.RegisterUnaryTransitCallback(w2.Wrap())

	// Add Capacity constraint.
	routing.AddDimensionWithVehicleCapacity(capacityCallbackIndex,
		0,                      // null capacity slack
		data.vehicleCapacities, // vehicle maximum capacities
		true,                   // start cumul to zero
		"Capacity")

	// Add time window constraints for each location except depot and 'copy' the
	// slack var in the solution object (aka Assignment) to print it.
	for i := 1; i < len(data.timeWindows); i++ {
		index := manager.NodeToIndex(i)
		timeDimension.CumulVar(index).SetRange(data.timeWindows[i][0],
			data.timeWindows[i][1])
		routing.AddToAssignment(timeDimension.SlackVar(index))
	}
	// Add time window constraints for each vehicle start node and 'copy' the
	// slack var in the solution object (aka Assignment) to print it.
	// Warning: Slack var is not defined for vehicle end nodes and should not be
	// added to the assignment
	for i := 0; i < len(data.vehicleCapacities); i++ {
		index := routing.Start(i)
		timeDimension.CumulVar(index).SetRange(data.timeWindows[0][0],
			data.timeWindows[0][1])
		routing.AddToAssignment(timeDimension.SlackVar(index))
	}

	// Instantiate route start and end times to produce feasible times.
	for i := 0; i < len(data.vehicleCapacities); i++ {
		routing.AddVariableMinimizedByFinalizer(
			timeDimension.CumulVar(routing.Start(i)))
		routing.AddVariableMinimizedByFinalizer(
			timeDimension.CumulVar(routing.End(i)))
	}

	// Define Transportation Requests.
	solver := routing.Solver()
	for _, request := range data.pickupDeliveries {
		pickupIndex := manager.NodeToIndex(request[0])
		deliveryIndex := manager.NodeToIndex(request[1])
		routing.AddPickupAndDelivery(pickupIndex, deliveryIndex)
		solver.AddConstraint(
			solver.MakeEquality(
				constraintsolver.SwigcptrIntExpr(routing.VehicleVar(pickupIndex).Swigcptr()),
				constraintsolver.SwigcptrIntExpr(routing.VehicleVar(deliveryIndex).Swigcptr())))
		solver.AddConstraint(
			solver.MakeLessOrEqual(
				constraintsolver.SwigcptrIntExpr(timeDimension.CumulVar(pickupIndex).Swigcptr()),
				constraintsolver.SwigcptrIntExpr(timeDimension.CumulVar(deliveryIndex).Swigcptr())))
	}

	// Define Allowed Vehicles per node.
	for i := 0; i < len(data.allowedVehicles); i++ {
		// If all vehicles are allowed, do not add any constraint
		if len(data.allowedVehicles[i]) == 0 || len(data.allowedVehicles[i]) == len(data.vehicleCapacities) {
			continue
		}
		nodeIndex := manager.NodeToIndex(i)
		routing.SetAllowedVehiclesForIndex(data.allowedVehicles[i], nodeIndex)
	}

	i := 0                           // Current solution count
	max := 15                        // Max # of solutions
	objectives := make([]int64, max) // Keep track of solution objective costs

	// Define callback function to handle solutions as they are found.
	p := func() {
		o := routing.CostVar().Value()
		if i == 0 || o < objectives[i-1] {
			fmt.Printf("Found Solution! Objective: %v\n", o)
			objectives[i] = o
			printSolution(data, manager, routing, nil)
			i++
		}
		// Stop search after max solutions are found.
		if i >= max {
			routing.Solver().FinishCurrentSearch()
		}
	}
	wp := constraintsolver.NewGoAtSolutionCallbackWrapper(p)
	routing.AddAtSolutionCallback(wp.Wrap())

	// wrapper deletion for deferred cleanup
	c := func() {
		wp.Delete()
		w2.Delete()
		w.Delete()
	}

	return routing, data, manager, c
}
