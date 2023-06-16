func addFiles(cmd *cobra.Command, args []string) {
	dir := args[0]
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Printf("Backup directory is empty")
		return
	}
	numBlocks := calculateNumBlocks(fileInfo.Size(), len(cfg.Servers))
	blockSize := calculateBlockSize(fileInfo.Size(), numBlocks)
	for _, file := range files {
		if file.Mode().IsRegular() {
			addFile(filepath.Join(dir, file.Name()), &cfg, numBlocks, blockSize)
		}
	}
}

func addFile(path string, cfg *config.Config, numBlocks int64, blockSize int64) {
	filePath := path
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("Error getting file information: %v\n", err)
		os.Exit(1)
	}
	checkSum, err := fileChecksum(file)
	if err != nil {
		fmt.Printf("Error getting file checksum: %v\n", err)
		os.Exit(1)
	}
	blocks := make([]config.FileBlock, 0)

	for i := 0; i < numBlocks; i++ {
		offset := int64(i) * blockSize
		blockSize := blockSize
		if i == numBlocks-1 {
			blockSize = fileInfo.Size() - offset
		}
		checkSum, err := blockChecksum(file, offset, blockSize)
		if err != nil {
			fmt.Printf("Error getting block checksum: %v\n", err)
			os.Exit(1)
		}
		blockPath := filepath.Join(blockPath, fmt.Sprintf("%s.%d", fileName[len(fileName)-1], i))
		fileName := strings.Split(filePath, "/")
		block := config.FileBlock{
			Id:       i,
			Size:     int(blockSize),
			Checksum: checkSum,
			Location: blockPath,
		}
		blocks = append(blocks, block)
		err = os.MkdirAll(filepath.Dir(blockPath), 0755)
		if err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			os.Exit(1)
		}
		err = saveBlockToFile(file, offset, blockSize, blockPath)
		if err != nil {
			fmt.Printf("Error saving block to file: %v\n", err)
			os.Exit(1)
		}
	}
	fileInf := config.FileInfo{
		Filename:  filepath.Base(filePath),
		Size:      fileInfo.Size(),
		BlockSize: blockSize,
		Checksum:  checkSum,
		Blocks:    blocks,
		Date:      time.Now().Format("2006-01-02 15:04:05"),
	}
	cfg.Files = append(cfg.Files, fileInf)
	err = saveConfig(*cfg)
	if err != nil {
		fmt.Printf("Error saving config file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("File info and blocks saved successfully.")
}

func loadConfig() (config.Config, error) {
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return config.Config{}, fmt.Errorf("error opening config file: %v", err)
	}
	defer configFile.Close()

	var cfg config.Config
	err = json.NewDecoder(configFile).Decode(&cfg)
	if err != nil {
		return config.Config{}, fmt.Errorf("error decoding config file: %v", err)
	}

	return cfg, nil
}

func fileChecksum(file *os.File) (string, error) {
	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	checksum := hash.Sum(nil)

	return hex.EncodeToString(checksum), nil
}

func blockChecksum(file *os.File, offset, size int64) (string, error) {
	hash := sha256.New()

	if _, err := file.Seek(offset, 0); err != nil {
		return "", err
	}

	buffer := make([]byte, size)
	if _, err := file.Read(buffer); err != nil {
		return "", err
	}

	if _, err := hash.Write(buffer[:size]); err != nil {
		return "", err
	}

	checksum := hash.Sum(nil)

	return hex.EncodeToString(checksum), nil
}

func saveBlockToFile(file *os.File, offset, size int64, filePath string) error {
	blockFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer blockFile.Close()
	if _, err := file.Seek(offset, 0); err != nil {
		return err
	}

	actualSize := size
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()
	if offset+size > fileSize {
		actualSize = fileSize - offset
	}
	_, err = io.CopyN(blockFile, file, actualSize)
	if err != nil {
		return err
	}

	if actualSize < size {
		remainingSize := size - actualSize
		zeroBytes := make([]byte, remainingSize)
		_, err = blockFile.Write(zeroBytes)
		if err != nil {
			return err
		}
	}

	return nil
}

func saveConfig(cfg config.Config) error {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %v", err)
	}

	err = ioutil.WriteFile(configFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	fmt.Printf("Config saved to %s\n", configFilePath)
	return nil
}
