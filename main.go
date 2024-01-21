package main

func main() {
	// flag.Parse()
	// if *flagShowHelp {
	// 	flag.Usage()
	// 	return 0
	// }

	// if *flagVersion {
	// 	fmt.Printf("%s\n", getVersion())
	// 	return 0
	// }

	// if *flagWaitHost != "" {
	// 	err := waitHost(*flagWaitHost, *flagWaitHostTimeout)
	// 	if err != nil {
	// 		fmt.Fprintf(os.Stderr, "%v\n", err)
	// 		return 1
	// 	}
	// 	return 0
	// }

	// rand.Seed(time.Now().UnixNano())

	// err := mainContext.Init()
	// if err != nil {
	// 	return 1
	// }
	// defer mainContext.Fin()

	// stopChan := make(chan bool, 1)

	// DebugPrintf("p2pcp server starting.")

	// server := CreateHttpServer(&mainContext)
	// listener, err := net.Listen("tcp", server.Addr)
	// if err != nil {
	// 	ErrorPrintf("p2pcp server Listen failed.: %v", err)
	// 	return 1
	// }

	// var returnCode atomic.Int32
	// var wg sync.WaitGroup

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()

	// 	err := server.Serve(listener)
	// 	if err == http.ErrServerClosed {
	// 		err = nil
	// 	}
	// 	if err != nil {
	// 		ErrorPrintf("p2pcp server has failed.: %v", err)
	// 		stopChan <- true
	// 		returnCode.Store(1)
	// 		return
	// 	}

	// 	if mainContext.PrintStatistics {
	// 		fmt.Printf("[%v] [Total] transmitted_bytes = %v\n", mainContext.Hostname, mainContext.BytesTransmittedTotal.Load())
	// 	}
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()

	// 	if mainContext.ServeOnly {
	// 		return
	// 	}

	// 	err := downloadFiles(&mainContext)
	// 	if err != nil {
	// 		ErrorPrintf("downloadFiles failed.: %v", err)
	// 		stopChan <- true
	// 		returnCode.Store(1)
	// 		return
	// 	}

	// 	if *flagExitComplete {
	// 		stopChan <- true
	// 	}
	// }()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()

	// 	signalChan := make(chan os.Signal, 1)
	// 	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 	select {
	// 	case sig := <-signalChan:
	// 		ErrorPrintf("received signal %v", sig)
	// 	case <-stopChan:
	// 	}

	// 	mainContext.WantStopDownload.Store(true)

	// 	err := CloseServer(&mainContext, server)
	// 	if err != nil {
	// 		ErrorPrintf("Shutting down p2pcp server failed.:%v", err)
	// 		returnCode.Store(1)
	// 	}
	// }()

	// wg.Wait()

	// ErrorPrintf("p2pcp server stopped.")

	// return int(returnCode.Load())
}
