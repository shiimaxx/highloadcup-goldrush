package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type ErrorRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type BalanceRes struct {
	Balance int `json:"balance"`
	Wallet  int `json:"wallet"`
}

type LicenseListRes LicenseList

type LicenseRes struct {
	ID         int `json:"id"`
	DigAllowed int `json:"digAllowed"`
	DigUsed    int `json:"digUsed"`
}

type ExploreRes struct {
	Area   Area `json:"area"`
	Amount int  `json:"amount"`
}

type DigRes []string

type CashRes []int

type Error struct {
	Code    int
	Message string
}

type Balance struct {
	Balance int
	Wallet  int
}

type License struct {
	ID         int
	DigAllowed int
	DigUsed    int
}

type LicenseList []License

type Area struct {
	PosX  int
	PosY  int
	SizeX int
	SizeY int
}

func NewArea(posX, posY int) *Area {
	return &Area{
		posX,
		posY,
		1,
		1,
	}
}

type Report struct {
	Area   Area
	Amount int
}

type Dig struct {
	LicenseID int
	PosX      int
	PosY      int
	Depth     int
}

type Treasure struct {
	Priority  int
	Treasures []string
}

type TreasureList []Treasure

type Explore struct {
	Priority int
	Area     *Area
	Amount   int
}

type Client struct {
	BaseURL string
	Client  *http.Client
	License *License
}

func (c *Client) PostLicense(coin []int) (*License, error) {
	log.Println("debug: license")

	url := fmt.Sprintf("%s/licenses", c.BaseURL)

	b, err := json.Marshal(coin)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("debug: status code is ", resp.StatusCode)
		return nil, nil 
	}

	licenseRes := &LicenseRes{}
	if err := json.NewDecoder(resp.Body).Decode(licenseRes); err != nil {
		return nil, err
	}

	return &License{
		ID:         licenseRes.ID,
		DigAllowed: licenseRes.DigAllowed,
		DigUsed:    licenseRes.DigUsed,
	}, nil
}

func (c *Client) PostDig(dig *Dig) (*Treasure, error) {
	log.Println("dig")

	url := fmt.Sprintf("%s/dig", c.BaseURL)

	b, err := json.Marshal(dig)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("debug: status code is ", resp.StatusCode)
		return nil, nil 
	}

	digRes := &DigRes{}
	if err := json.NewDecoder(resp.Body).Decode(digRes); err != nil {
		return nil, err
	}

	return &Treasure{
		Priority:  0,
		Treasures: *digRes,
	}, nil
}

func (c *Client) PostCash(treasure string) (*CashRes, error) {
	log.Println("cash")

	url := fmt.Sprintf("%s/cash", c.BaseURL)

	b, err := json.Marshal(treasure)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("debug: status code is ", resp.StatusCode)
		return nil, nil 
	}

	cashRes := &CashRes{}
	if err := json.NewDecoder(resp.Body).Decode(cashRes); err != nil {
		return nil, err
	}

	return cashRes, nil
}

func (c *Client) PostExplore(area *Area) (*Explore, error) {
	log.Println("explore")

	url := fmt.Sprintf("%s/explore", c.BaseURL)

	b, err := json.Marshal(area)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("debug: status code is ", resp.StatusCode)
		return nil, nil 
	}

	exploreRes := &ExploreRes{}
	if err := json.NewDecoder(resp.Body).Decode(exploreRes); err != nil {
		return nil, err
	}

	return &Explore{
		Priority: 0,
		Area:     area,
		Amount:   exploreRes.Amount,
	}, nil
}

func (c *Client) UpdateLicense(ch chan *License) error {
	c.License = <- ch

	return nil
}

func Game(client *Client) error {
	ExploreChan := make(chan *Explore, 100)
	LicenseChan := make(chan *License, 100)
	ErrChan := make(chan error)

	go func() {
		for {
			var coins []int

			license, err := client.PostLicense(coins)
			if err != nil {
				ErrChan <- err
			}

			LicenseChan <- license
		}
	}()

	go func() {
		for {
			result := <-ExploreChan

			depth := 1
			left := result.Amount

			for depth <= 10 && left > 0 {
				for client.License == nil || client.License.DigUsed >= client.License.DigAllowed {
					if err := client.UpdateLicense(LicenseChan); err != nil {
						ErrChan <- err
					}
				}

				dig := &Dig{
					LicenseID: client.License.ID,
					PosX:      result.Area.PosX,
					PosY:      result.Area.PosY,
					Depth:     depth,
				}
				treasures, err := client.PostDig(dig)
				if err != nil {
					ErrChan <- err
				}

				client.License.DigUsed += 1
				depth += 1

				if treasures != nil {
					for _, treasure := range treasures.Treasures {
						res, err := client.PostCash(treasure)
						if err != nil {
							ErrChan <- err
						}

						if res != nil {
							left -= 1
						}
					}
				}
			}
		}
	}()

	for x := 0; x < 3500; x++ {
		for y := 0; y < 3500; y++ {
			select {
			case err := <- ErrChan:
				return err
			default:
				area := NewArea(x, y)
				result, err := client.PostExplore(area)
				if err != nil {
					return err
				}

				if result == nil || result.Amount < 1 {
					log.Printf("debug: skip: %+v\n", result)
					continue
				}
				ExploreChan <- result
			}
		}
	}

	return nil
}

func main() {
	address := os.Getenv("ADDRESS")
	baseURL := fmt.Sprintf("http://%s:%d", address, 8000)

	log.Fatal(Game(&Client{baseURL, http.DefaultClient, &License{}}))
}
