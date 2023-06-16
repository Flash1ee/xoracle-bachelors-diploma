func restoreFiles(cmd *cobra.Command, args []string) error {
    var err error
    cfg, err = loadConfig()
    if err != nil {
        log.Fatal("error load config")
    }
    if err := Restore(&cfg); err != nil {
        return fmt.Errorf("backup failed: %v", err)
    }

    fmt.Println("Restore completed successfully.")
    return nil
}
func Restore(cfg *config.Config, isTest bool) error {
    badBlocks := make(map[string]config.FileBlock, 2)
    for _, val := range cfg.Files {
        for _, block := range val.Blocks {
            _, err := os.Stat(block.Location)
            if os.IsNotExist(err) {
                badBlocks[block.Location] = block
            } else if err != nil {
                return fmt.Errorf("ошибка при проверке файла блока: %v", err)
            } else {
                dir, fileName := filepath.Split(block.Location)
                err = VerifyBlockChecksum(cfg, dir, fileName, int64(block.Size))
                if err != nil {
                    badBlocks[block.Location] = block
                }
            }
        }
    }
    if len(badBlocks) == 0 {
        fmt.Println("all blocks are correct")
        return nil
    }
    for _, val := range cfg.SuperBlocks {
        if len(val.Blocks) == 1 {
            if _, ok := badBlocks[val.Blocks[0].Location]; ok {
                blockBytes, err := readSuperblock(val.Location)
                if err != nil {
                    log.Println(fmt.Sprintf("error read log %s", val.Location))
                    return err
                }
                err = SaveBytesToFile(blockBytes.Blocks, val.Blocks[0].Location)
                if err != nil {
                    log.Println(fmt.Sprintf("error save recovered block: %v", val.Blocks[0].Location))
                    return err
                }
            }
            continue
        }
        // New block save here
        var badBlockLocation string
        // XOR superblock with recoverBlock
        recoverBlock := config.FileBlock{}
        block1 := val.Blocks[0]
        block2 := val.Blocks[1]
        _, ok := badBlocks[block1.Location]
        if ok {
            badBlockLocation = block1.Location
            recoverBlock = block2
        } else {
            _, ok2 := badBlocks[block2.Location]
            if ok2 {
                badBlockLocation = block2.Location
                recoverBlock = block1
            } else {
                continue
            }
        }
        recoverBlockBytes, err := readBlock(recoverBlock.Location)
        if err != nil {
            log.Println(fmt.Sprintf("error read log %s", recoverBlock.Location))
            return err
        }
        xoredBlockBytes, err := readSuperblock(val.Location)
        if err != nil {
            log.Println(fmt.Sprintf("error read log %s", recoverBlock.Location))
            return err
        }
        xorRecoverBlockBytes := xorBlock(recoverBlockBytes, xoredBlockBytes.Blocks)[:val.BlockSize]
        newChecksum, err := bytesChecksum(xorRecoverBlockBytes)
        if err != nil {
            return fmt.Errorf("error get checksum: %w", err)
        }
        _, badBlockName := filepath.Split(badBlockLocation)
        expectedChecksum := getBlockChecksum(badBlockName, cfg)

        if expectedChecksum != newChecksum {
            return fmt.Errorf("checksum %s mismatch: expected = %s, actual = %s", badBlockName, expectedChecksum, newChecksum)
        }
        delete(badBlocks, badBlockLocation)

        err = SaveBytesToFile(xorRecoverBlockBytes, badBlockLocation)
        if err != nil {
            log.Println(fmt.Sprintf("error save recovered block: %v", badBlockLocation))
            return err
        }
    }
    return nil
}
func xorBlock(dst, src []byte) {
    for i := 0; i < len(dst); i++ {
        dst[i] ^= src[i]
    }
}

func removePadding(block []byte) []byte {
    cnt := 0
    i := len(block) - 1
    for ; i >= 0; i-- {
        if block[i] != 0 {
            break
        }
        cnt += 1
    }
    return block[:i+1]
}

func getLastWordWithoutExtension(filePath string) string {
    base := filepath.Base(filePath)
    name := strings.TrimSuffix(base, filepath.Ext(base))
    words := strings.Fields(name)
    if len(words) > 0 {
        return words[len(words)-1]
    }
    return ""
}

func readSuperblock(filePath string) (*superBlock, error) {
    // Открываем файл суперблока для чтения
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open superblock file %s: %v", filePath, err)
    }
    defer file.Close()

    // Читаем данные из файла
    data, err := ioutil.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("failed to read superblock file %s: %v", filePath, err)
    }

    // Декодируем данные из JSON
    var sb superBlock
    err = json.Unmarshal(data, &sb)
    if err != nil {
        return nil, fmt.Errorf("failed to decode superblock from file %s: %v", filePath, err)
    }

    return &sb, nil
}

func saveRestoredBlocks(restoreDir, filename string, blocks []byte) error {
    filePath := filepath.Join(restoreDir, filename)
    err := ioutil.WriteFile(filePath, blocks, os.ModePerm)
    if err != nil {
        return fmt.Errorf("failed to save restored blocks to file %s: %v", filePath, err)
    }
    return nil
}
