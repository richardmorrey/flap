package flap

import  (
	"github.com/richardmorrey/flap/pkg/flap/db"
	"encoding/binary"
	"bytes"
	"os"
	"bufio"
	"encoding/csv"
	"io"
	"strconv"
)

type ICAOCode [4]byte

func (self *ICAOCode) ToString() string {
	return string(self[:])
}

type Airport struct {
	Code ICAOCode
	Loc  LatLon
}

func NewICAOCode(codestring string) ICAOCode {
	var code ICAOCode
	copy(code[:],codestring)
	return code
}

type Airports struct {
	table db.Table
}

// NewAirports is the factory function for Airports
func NewAirports(database db.Database)  *Airports{
	a := new(Airports)
	var err error
	a.table,err = database.OpenTable(airportsTableName)
	if err != nil {
		a.table,err = database.CreateTable(airportsTableName)
		if err != nil {
			return nil
		}
	}
	return a
}

// Drops airports table from given database
func dropAirports(database db.Database) error {
	return database.DropTable(airportsTableName)
}

// LoadAirports populates a table "airports" in the given database from a csv file
// CSV file must be fomatted as per "airpots.dat" file from https://openflights.org/data.html.
// If the table doesnt exist it is created.
// If the table already exists it is emptied first.
// Each entry table is keyed by ICAOCode and holds the latitude and longitude of the airport
const airportsTableName="airports"
func (self  *Airports) LoadAirports(filepath string) error {
	
	// Open and iterate through CSV file
	csvFile, err := os.Open(filepath)
	if (err != nil) {
		return err
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else
		if err != nil {
			return err
		}
		
		// Add record as binary
		var loc LatLon
		var buf bytes.Buffer
		loc.Lat,err=strconv.ParseFloat(line[6],64)
		if err != nil {
			return err
		}
		loc.Lon,err=strconv.ParseFloat(line[7],64)
		if err != nil {
			return err
		}
		binary.Write(&buf, binary.LittleEndian,loc)
		err = self.table.Put([]byte(line[5]),buf.Bytes())
		if (err != nil) {
			return err
		} 
	}
	return nil
}

// GetAirport creates and returns an Airport struct populated from the appropriate
// record in the current "airports" table. If that table doesnt exist or
// doesnt contain an entry for the given ICAOCode an empty Airport and error is returned.
func (self *Airports) GetAirport(code  ICAOCode) (Airport,error) {
	
	// Retrieve raw record
	var airport Airport
	if  self.table == nil {
		return Airport{},ETABLENOTOPEN
	}
	blob,err := self.table.Get(code[:])
	if (err != nil) {
		return Airport{},err
	}

	// Deserialize to struct
	buf := bytes.NewReader(blob)
	err = binary.Read(buf,binary.LittleEndian,&airport.Loc)
	if (err !=nil) {
		return Airport{},err
	}
	airport.Code=code
	return airport,nil
}

