func Backup(configData *config.Config, isTest bool) error {
    for _, server := range configData.Servers {
        err := os.MkdirAll(server.Path, os.ModePerm)
        if err != nil {
            return fmt.Errorf("failed to create directory %s: %v", server.Path, err)
        }
    }
    serverIndex := 0
    for _, file := range configData.Files {
        blocks := readBlocks(file.Blocks)
        superblockData := xorBlocks(blocks, int(file.BlockSize))
        for idx, val := range superblockData {
            serverPath := configData.Servers[serverIndex].Path
            backupDir := filepath.Join(serverPath)
            err := os.MkdirAll(backupDir, os.ModePerm)
            if err != nil {
                return fmt.Errorf("failed to create Backup directory: %v", err)
            }

            superblock := superBlock{
                ID:     idx,
                Blocks: val.Data,
            }
            saveLocation := fmt.Sprintf("%s.%d", file.Filename, idx)
            err = writeSuperblock(backupDir, saveLocation, &superblock)
            if err != nil {
                return fmt.Errorf("failed to write superblock for %s: %v", file.Filename, err)
            }
            updateBlockLocationInConfig(configData, backupDir, file.Filename, superblockBlocks)
            serverIndex = (serverIndex + 1) % len(configData.Servers)
        }

    }
    configPath := configFilePath
    if isTest {
        configPath = configFileTestPath
    }
    err := saveConfig(*configData, configPath)
    if err != nil {
        return fmt.Errorf("failed to save configData: %v", err)
    }

    return nil
}

func xorBlocks(blocks []BlockInfo, fileBlockSize int) []XorData {
    if len(blocks) == 1 {
        return []XorData{
            {
                Size:        blocks[0].Size,
                CountBlocks: 1,
                Location: []string{
                    blocks[0].Location,
                },
                Data: blocks[0].Data,
            },
        }
    }
    xorBlocks := make([]XorData, 0, 1)
    if len(blocks) > 1 {
        xorBlocks = make([]XorData, int(math.Ceil(float64(len(blocks))/2.0)))
    }
    for i := 0; i < len(blocks)-1; i++ {
        block1 := blocks[i]
        block2 := blocks[i+1]
        if block2.Size != fileBlockSize {
            block2.Data = addPadding(block2.Data, fileBlockSize)
        }
        xorBlock := XorData{
            Size:        block2.Size,
            CountBlocks: 2,
            Location: []string{
                block1.Location,
                block2.Location,
            },
            Data: make([]byte, fileBlockSize),
        }
        for j := 0; j < fileBlockSize; j++ {
            xorBlock.Data[j] = block1.Data[j] ^ block2.Data[j]
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
