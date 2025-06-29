package main

import (
	"aavev3-raw-balances-collector/internal/balances"
	"aavev3-raw-balances-collector/internal/blockfinder"
	"aavev3-raw-balances-collector/internal/datalab"
	"aavev3-raw-balances-collector/internal/pool"
	"aavev3-raw-balances-collector/internal/utils"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	start := os.Getenv("START_DATE")
	stop := os.Getenv("END_DATE")
	accessKeyID := os.Getenv("ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("SECRET_ACCESS_KEY")
	provider := os.Getenv("PROVIDER")
	UIPoolDataProviderAddress := os.Getenv("POOL_DATA_PROVIDER")

	startDate, err := time.Parse("2006-01-02", start)
	if err != nil {
		panic(err)
	}
	stopDate, err := time.Parse("2006-01-02", stop)
	if err != nil {
		panic(err)
	}

	for day := startDate; day.Before(stopDate); day = day.AddDate(0, 0, 1) {
		err := DailyEtl(day, UIPoolDataProviderAddress, accessKeyID, secretAccessKey, provider)
		if err != nil {
			panic(err)
		}
	}
}

func DailyEtl(day time.Time, UIPoolDataProviderAddress, accessKeyID, secretAccessKey, provider string) error {
	fmt.Printf("Starting job for day %v\n", day)

	endpoint := "minio-simple.lab.groupe-genes.fr"
	bucket := "projet-datalab-group-jprat"
	day_str := fmt.Sprint(day)[:10]
	input_path := "aavev3-raw-datasource/daily-decoded-events/decoded_events_snapshot_date=" + day_str + "/all_active_users.json"
	output_path := "aavev3-raw-datasource/daily-users-balances/users_balances_snapshot_date=" + day_str + "/"

	fmt.Println("STEP 1 - Connecting to provider...")
	client, err := ethclient.Dial(provider)
	if err != nil {
		panic(err)
	}

	fmt.Println("STEP 2 - Setting pool contract...")
	poolCtr, err := pool.NewPool(common.HexToAddress("0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2"), client)
	if err != nil {
		panic(err)
	}

	fmt.Println("STEP 3 - Extracting users to query...")
	users, err := datalab.ReadActiveUsers(endpoint, bucket, input_path, accessKeyID, secretAccessKey)
	if err != nil {
		panic(err)
	}

	fmt.Println("STEP 4 - Finding end block of the day...")
	// Finding block of the end of the day
	dayTmstp := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	dayEndTmstp := dayTmstp.AddDate(0, 0, 1)

	references := utils.GetBlockReferences()
	refBlockNumber := references[time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, time.UTC).Unix()]

	_, endDayBlock, err := blockfinder.FindClosestBlocks(client, uint64(dayEndTmstp.Unix()), refBlockNumber, 7000)
	if err != nil {
		return err
	}

	// endDayBlock.Set(big.NewInt(22796291))

	fmt.Printf("   --> INFO: BlockNumber is %v\n", endDayBlock)
	fmt.Println("STEP 5 - Collecting tokens info...")
	tokens, err := balances.CollectAaveTokens(client, poolCtr, endDayBlock)
	if err != nil {
		return err
	}

	fmt.Println("STEP 6 - Collecting reserves data...")
	reservesData, err := balances.CollectReservesData(poolCtr, tokens, endDayBlock)
	if err != nil {
		return err
	}

	fmt.Println("STEP 7 - Collecting users balances...")
	// dataProvAddress := common.HexToAddress(UIPoolDataProviderAddress)
	// dataProviderCtr, err := dataprovider.NewDataprovider(dataProvAddress, client)
	// if err != nil {
	// 	return err
	// }
	// usersBalances, err := balances.CollectAllUsersBalances(users, poolCtr, dataProviderCtr, tokens, endDayBlock)
	// if err != nil {
	// 	return err
	// }

	// dataProviderCtr, err := prevdataprovider.NewPrevdataprovider(dataProvAddress, client)
	// if err != nil {
	// 	return err
	// }
	// usersBalances, err := balances.CollectAllUsersBalancesPrev(users, poolCtr, dataProviderCtr, tokens, endDayBlock)
	// if err != nil {
	// 	return err
	// }

	usersBalances, err := balances.CollectAllUsersBalancesManual(users, poolCtr, tokens, endDayBlock)
	if err != nil {
		return err
	}

	fmt.Println("STEP 8 - Generating and saving outputs...")
	datalab.SaveRecords(endpoint, accessKeyID, secretAccessKey, reservesData, bucket, output_path+"reserves_data.json")
	datalab.SaveRecords(endpoint, accessKeyID, secretAccessKey, usersBalances, bucket, output_path+"active_users_balances.json")
	fmt.Println("Done!")
	return nil
}
