package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/google/logger"
	"github.com/syndtr/goleveldb/leveldb"
	ldb_opt "github.com/syndtr/goleveldb/leveldb/opt"
)

var versionKey = []byte{0x00, 'V', 'E', 'R', 'S', 'I', 'O', 'N'}

// SetDBUpdaterReady is to setup the specific type of RemoteLatestChainFetcher and RemoteDBDownloader
func SetDBUpdaterReady(conf DBUpdaterConfig) (DBUpdater, error) {
	updaterConfig := conf.(DBUpdaterHTTPSConfig).GetConfig()
	log.Warning(updaterConfig.GetConfig().(DBUpdaterHTTPSConfig).APIEndpoint)
	httpUpdater := &DBUpdaterHTTPS{
		LatestChainInfoEndpoint: updaterConfig.(DBUpdaterHTTPSConfig).APIEndpoint,
		CurrentDBPath:           updaterConfig.(DBUpdaterHTTPSConfig).CurrentDBPath,
		ZipSourcePath:           updaterConfig.(DBUpdaterHTTPSConfig).ZipSourcePath,
		ZipDestinationPath:      updaterConfig.(DBUpdaterHTTPSConfig).ZipDestinationPath,
	}
	// get the currentDBVersion
	_, _, err := httpUpdater.GetCurrentDBVersion()
	if err != nil {
		if !os.IsNotExist(err) {
			return httpUpdater, err
		}
	}
	latest, err := httpUpdater.GetLatestChain()
	if err != nil {
		return httpUpdater, err
	}
	if latest != nil {
		httpUpdater.Latest = *latest
	}

	return httpUpdater, nil
}

// GetCurrentDBVersion get current chainData version
func (r *DBUpdaterHTTPS) GetCurrentDBVersion() (mainnet int, testbet int, err error) {

	opt := &ldb_opt.Options{
		ErrorIfExist:   false,
		ErrorIfMissing: true,
		ReadOnly:       true,
	}

	db, err := leveldb.OpenFile(r.CurrentDBPath, opt)
	if nil != err {
		return 0, 0, err
	}

	versionValue, err := db.Get(versionKey, nil)
	if leveldb.ErrNotFound == err {
		return 0, 0, nil
	} else if nil != err {
		return 0, 0, err
	}
	if 4 != len(versionValue) {
		db.Close()
		log.Errorf("incompatible database version length: expected: %d  actual: %d", 4, len(versionValue))
		return 0, 0, ErrorIncompatibleVersionLength
	}
	version := int(binary.BigEndian.Uint32(versionValue))
	r.CurrentDBVer = version
	db.Close()
	log.Info("GetCurrentDBVersion Successfully")
	// TODO: need to do update testnet
	return version, 0, nil
}

// IsUpdated is to check if current databse is updated
func (r *DBUpdaterHTTPS) IsUpdated() (main bool, test bool) {
	if r.IsForceUpdate() {
		return false, true
	}
	if r.CurrentDBVer != 0 {
		latestVer, err := r.Latest.GetVerion()
		if err != nil {
			return false, false
		}
		if latestVer != r.CurrentDBVer {
			return false, false
		}
		return true, true
	} else {
		return false, true
	}
}

// GetLatestChain to get latestChainInfo from Retmote
func (r *DBUpdaterHTTPS) GetLatestChain() (*LatestChain, error) {
	resp, err := http.Get(r.LatestChainInfoEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var latestChain LatestChain
	err = json.Unmarshal(body, &latestChain)
	if err != nil {
		return nil, err
	}
	return &latestChain, err
}

// IsStartFromGenesis is to check if current databse is updated
func (r *DBUpdaterHTTPS) IsStartFromGenesis() bool {
	if true == r.Latest.FromGenesis {
		return true
	}
	return false
}

// IsForceUpdate is to check if current databse is updated
func (r *DBUpdaterHTTPS) IsForceUpdate() bool {
	if true == r.Latest.ForceUpdate {
		return true
	}
	return false
}

// UpdateToLatestDB Download latest and update the local database
func (r *DBUpdaterHTTPS) UpdateToLatestDB() error {
	mainnetUpdated, testnetUpdated := r.IsUpdated()
	if mainnetUpdated && testnetUpdated {
		log.Info("mainnet and testnet are all updated")
		return nil
	}
	if r.IsStartFromGenesis() { // Chain should start from genesis
		//Rename old Database
		renameErr := renameBitmarkdDB()
		if renameErr != nil {
			return renameErr
		}
		return nil
	}

	if !mainnetUpdated {
		err := r.downloadfile("mainnet")
		if err != nil {
			return err
		}
		err = renameBitmarkdDB()
		if err != nil {
			return err
		}
		err = unzip(r.ZipSourcePath, r.ZipDestinationPath)
		if err != nil {
			recoverErr := recoverBitmarkdDB()
			r.Latest = LatestChain{}
			return ErrCombind(err, recoverErr)
		}
		log.Warning("UpdateToLatestDB Successful")

		err = removeFile(r.ZipSourcePath)
		if err != nil { // nice to have so does not return error even it has error
			log.Warning("UpdateToLatestDB:remove zip file error:", err)
		}
	}

	// TODO:for testnet

	return nil
}

func (r *DBUpdaterHTTPS) downloadfile(network string) error {
	var downloadURL string
	if "testnet" == network {
		downloadURL = r.Latest.TestDataURL
	} else {
		downloadURL = r.Latest.DataURL
	}
	log.Warning("### downloadURL" + downloadURL)
	resp, err := http.Get(downloadURL)

	if err != nil {
		log.Error("Get file error:", err)
		return err
	}
	defer resp.Body.Close()
	// Create the file
	if 0 == len(r.ZipSourcePath) {
		baseDir, err := builDefaultVolumSrcBaseDir()
		if err != nil {
			return err
		}
		r.ZipSourcePath = filepath.Join(baseDir, "data", "snapshot.zip")
	}
	zipfile, err := os.Create(r.ZipSourcePath)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	// Write the body to file
	_, err = io.Copy(zipfile, resp.Body)
	if err != nil {
		return nil
	}

	return err
}
