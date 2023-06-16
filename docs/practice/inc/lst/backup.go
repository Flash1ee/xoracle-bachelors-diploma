func backup(config *config.Config) error {
	for _, server := range config.Servers {
		err := os.MkdirAll(server.Path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", server.Path, err)
		}
	}
	serverIndex := 0
	for id, file := range cfg.Files {
		xorData := xorBlocks(file.Blocks)
		backupDir := filepath.Join(cfg.Servers[serverIndex].Path)

		err := os.MkdirAll(backupDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create backup directory: %v", err)
		}
		// Create superblock
		superblock := superBlock{
			ID:     id,
			Blocks: xorData,
		}
		// Save superblock in file
		err = writeSuperblock(backupDir, file.Filename, &superblock)
		if err != nil {
			return fmt.Errorf("failed to write superblock for %s: %v", file.Filename, err)
		}
		// Save information about superblock to config
		updateBlockLocationInConfig(&cfg, backupDir, file.Filename, file.Blocks)
		// Calculate next server by Round-Robin
		serverIndex = (serverIndex + 1) % len(cfg.Servers)
	}
	// Save new config version 
	err := saveConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil

}

func xorBlocks(blocks []config.FileBlock) [][]byte {
	if len(blocks) == 0 {
		return nil
	}
	blocksAll := make([][]byte, 0, len(blocks))
	for _, block := range blocks {
		blockData, err := readBlock(block.Location)
		if err != nil {
			log.Fatal("error read block")
		}
		blocksAll = append(blocksAll, blockData)
	}
	blockSize := len(blocksAll[0])
	xorBlocks := make([][]byte, len(blocks))
	for i, block := range blocksAll {
		xorBlock := make([]byte, blockSize)
		if i == 0 {
			copy(xorBlock, block[:blockSize])
		} else {
			if i == len(blocksAll)-1 {
				block = addPadding(block, blockSize)
			}
			for j, b := range block {
				xorBlock[j] = b ^ blocksAll[i-1][j]
			}
		}
		xorBlocks[i] = xorBlock
	}
	return xorBlocks
}

func addPadding(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)
	fmt.Println("padLen=", padLen)
	padding := bytes.Repeat([]byte{0}, padLen)
	return append(data, padding...)
}

func readBlock(location string) ([]byte, error) {
	f, err := os.Open(location)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", location, err)
	}
	defer f.Close()
	blockData, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read block from file %s: %v", location, err)
	}
	return blockData, nil
}

func writeSuperblock(destDir string, filePath string, sb *superBlock) error {
	log.Println(fmt.Sprintf("filePath = %v", filePath))
	destFile, err := os.Create(filepath.Join(destDir, filepath.Base(filePath)+".sb"))
	if err != nil {
		return fmt.Errorf("failed to create superblock file for %s: %v", filePath, err)
	}
	defer destFile.Close()
	data, err := json.Marshal(sb)
	if err != nil {
		return err
	}
	if err = binary.Write(destFile, binary.LittleEndian, data); err != nil {
		return err
	}
	return nil
}

func updateBlockLocationInConfig(cfg *config.Config, backupDir string, fileName string, blocks []config.FileBlock) {
	cfg.SuperBlocks = append(cfg.SuperBlocks, config.Superblock{
		Id:        0,
		BlockSize: blocks[0].Size,
		Location:  backupDir + "/" + fileName + ".sb",
		Blocks:    blocks,
	})
}
