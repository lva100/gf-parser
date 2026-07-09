package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
)

// Константы для батчинга
const (
	BatchSize = 3000 // Количество записей в одном батче
)

// Root struct - корневой элемент
type GFDocument struct {
	XMLName xml.Name `xml:"DN_PATIENT"`
	ZGLV    ZGLV     `xml:"ZGLV"`
	ZAPs    []ZAP    `xml:"ZAP"`
}

// Заголовок файла
type ZGLV struct {
	Version  string `xml:"VERSION"`
	FileType string `xml:"FILE_TYPE"`
	Data     string `xml:"DATA"`
	Filename string `xml:"FILENAME"`
	RegionCD string `xml:"REGION_CD"`
	Period   string `xml:"PERIOD"`
	SDZ      int    `xml:"SD_Z"`
}

// Запись о застрахованном лице
type ZAP struct {
	DN_Patient_ID    string  `xml:"DN_PATIENT_ID"`
	ENP              string  `xml:"ENP"`
	W                int     `xml:"W"`
	DR               string  `xml:"DR"`
	SMO              string  `xml:"SMO"`
	Attach_MCode     string  `xml:"ATTACH_MCODE"`
	Attach_Date      string  `xml:"ATTACH_DATE"`
	SMO_Region_CD    string  `xml:"SMO_REGION_CD"`
	Group_RH_CD      int     `xml:"GROUP_RH_CD"`
	Group_RH_DS      string  `xml:"GROUP_RH_DS"`
	DN_Prvs          int     `xml:"DN_PRVS"`
	Group_RH_Profile string  `xml:"GROUP_RH_PROFILE"`
	Group_RH_Name    string  `xml:"GROUP_RH_NAME"`
	DN_Rule_In_Name  string  `xml:"DN_RULE_IN_NAME"`
	DN_GIS           DN_GIS  `xml:"DN_GIS"`
	DN_LIST          DN_LIST `xml:"DN_LIST"`
	DN_PLAN          DN_PLAN `xml:"DN_PLAN"`
	Insert_DTTM      string  `xml:"INSERT_DTTM"`
	Update_DTTM      string  `xml:"UPDATE_DTTM"`
}

// Данные ГИС ОМС
type DN_GIS struct {
	Trigger_Schet_Filename string `xml:"TRIGGER_SCHET_FILENAME"`
	Trigger_Schet_Code     string `xml:"TRIGGER_SCHET_CODE"`
	Trigger_Nschet         string `xml:"TRIGGER_NSCHET"`
	Trigger_Dschet         string `xml:"TRIGGER_DSCHET"`
	Trigger_Idcase         string `xml:"TRIGGER_IDCASE"`
	Trigger_SL_Id          string `xml:"TRIGGER_SL_ID"`
	Trigger_SL_Nhistory    string `xml:"TRIGGER_SL_NHISTORY"`
	Trigger_DS_CD          string `xml:"TRIGGER_DS_CD"`
	Trigger_MCode          string `xml:"TRIGGER_MCODE"`
	Trigger_DT             string `xml:"TRIGGER_DT"`
}

// Результат сверки списка ЗЛ
type DN_LIST struct {
	DN_List_Period_CD     int    `xml:"DN_LIST_PERIOD_CD"`
	DN_List_Filename      string `xml:"DN_LIST_FILENAME"`
	CODE_L                string `xml:"CODE_L"`
	DN_List_Result_Code   string `xml:"DN_LIST_RESULT_CODE"`
	DN_List_Date_Checking string `xml:"DN_LIST_DATE_CHEKING"`
	DN_List_Result_Descr  string `xml:"DN_LIST_RESULT_DESCR"`
}

// Результат сверки план-графика
type DN_PLAN struct {
	DN_Plan_Period        string `xml:"DN_PLAN_PERIOD"`
	DN_Plan_Filename      string `xml:"DN_PLAN_FILENAME"`
	CODE_P                string `xml:"CODE_P"`
	DN_Plan_Result_Code   int    `xml:"DN_PLAN_RESULT_CODE"`
	DN_Plan_Date_Checking string `xml:"DN_PLAN_DATE_CHEKING"`
	DN_Plan_Result_Descr  string `xml:"DN_PLAN_RESULT_DESCR"`
}

// Структура для батчевой вставки
type GFRecord struct {
	// Заголовок
	Version      string
	FileType     string
	DataDate     interface{}
	FileName     string
	RegionCode   string
	Period       string
	RecordsCount int

	// Данные ZAP
	DN_Patient_ID    string
	ENP              string
	GenderCode       int
	BirthDate        interface{}
	SMO_Code         string
	Attach_MCode     string
	Attach_Date      interface{}
	SMO_Region_Code  string
	Group_RH_Code    int
	Group_RH_DS      string
	DN_Prvs          int
	Group_RH_Profile string
	Group_RH_Name    string
	DN_Rule_In_Name  string

	// Данные DN_GIS
	Trigger_Schet_Filename string
	Trigger_Schet_Code     string
	Trigger_Nschet         string
	Trigger_Dschet         interface{}
	Trigger_Idcase         string
	Trigger_SL_Id          string
	Trigger_SL_Nhistory    string
	Trigger_DS_CD          string
	Trigger_MCode          string
	Trigger_DT             interface{}

	// Данные DN_LIST
	DN_List_Period_CD     int
	DN_List_Filename      string
	CODE_L                string
	DN_List_Result_Code   string
	DN_List_Date_Checking interface{}
	DN_List_Result_Descr  string

	// Данные DN_PLAN
	DN_Plan_Period        interface{}
	DN_Plan_Filename      string
	CODE_P                string
	DN_Plan_Result_Code   int
	DN_Plan_Date_Checking interface{}
	DN_Plan_Result_Descr  string

	// Служебные
	Insert_DTTM string
	Update_DTTM string
}

// Конфигурация подключения к БД
type DBConfig struct {
	Server   string
	Port     int
	User     string
	Password string
	Database string
}

// Парсер с батчингом
type GFParser struct {
	db          *sql.DB
	batchSize   int
	totalSaved  int
	totalErrors int
}

func NewGFParser(config DBConfig) (*GFParser, error) {
	connString := fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s;encrypt=disable",
		config.Server, config.Port, config.User, config.Password, config.Database)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Настройка пула соединений для больших объёмов
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return &GFParser{
		db:         db,
		batchSize:  BatchSize,
		totalSaved: 0,
	}, nil
}

func (p *GFParser) Close() {
	if p.db != nil {
		p.db.Close()
	}
}

// Парсинг XML файла с батчингом
func (p *GFParser) ParseFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	var doc GFDocument
	err = xml.Unmarshal(data, &doc)
	if err != nil {
		return fmt.Errorf("ошибка парсинга XML: %w", err)
	}

	log.Printf("Файл: %s, Версия: %s, Кол-во записей: %d",
		doc.ZGLV.Filename, doc.ZGLV.Version, len(doc.ZAPs))

	// Обработка батчами
	totalRecords := len(doc.ZAPs)
	for i := 0; i < totalRecords; i += p.batchSize {
		end := i + p.batchSize
		if end > totalRecords {
			end = totalRecords
		}

		batch := doc.ZAPs[i:end]

		// Конвертируем в записи для БД
		records := make([]GFRecord, 0, len(batch))
		for _, zap := range batch {
			record := p.convertToRecord(zap, doc.ZGLV)
			records = append(records, record)
		}

		// Сохраняем батч
		if err := p.saveBatch(records); err != nil {
			log.Printf("Ошибка сохранения батча (записи %d-%d): %v", i+1, end, err)
			p.totalErrors += len(batch)
			continue
		}

		p.totalSaved += len(batch)
		log.Printf("Обработано записей: %d из %d (успешно: %d, ошибок: %d)",
			end, totalRecords, p.totalSaved, p.totalErrors)
	}

	log.Printf("Завершено. Всего записей: %d, Сохранено: %d, Ошибок: %d",
		totalRecords, p.totalSaved, p.totalErrors)
	return nil
}

// Конвертация ZAP в GFRecord
func (p *GFParser) convertToRecord(zap ZAP, header ZGLV) GFRecord {
	now := time.Now().Format("2006-01-02 15:04:05")

	return GFRecord{
		// Заголовок
		Version:      header.Version,
		FileType:     header.FileType,
		DataDate:     parseDate(header.Data),
		FileName:     header.Filename,
		RegionCode:   header.RegionCD,
		Period:       header.Period,
		RecordsCount: header.SDZ,

		// ZAP
		DN_Patient_ID:    zap.DN_Patient_ID,
		ENP:              zap.ENP,
		GenderCode:       zap.W,
		BirthDate:        parseDate(zap.DR),
		SMO_Code:         zap.SMO,
		Attach_MCode:     zap.Attach_MCode,
		Attach_Date:      parseDate(zap.Attach_Date),
		SMO_Region_Code:  zap.SMO_Region_CD,
		Group_RH_Code:    zap.Group_RH_CD,
		Group_RH_DS:      zap.Group_RH_DS,
		DN_Prvs:          zap.DN_Prvs,
		Group_RH_Profile: zap.Group_RH_Profile,
		Group_RH_Name:    zap.Group_RH_Name,
		DN_Rule_In_Name:  zap.DN_Rule_In_Name,

		// DN_GIS
		Trigger_Schet_Filename: zap.DN_GIS.Trigger_Schet_Filename,
		Trigger_Schet_Code:     zap.DN_GIS.Trigger_Schet_Code,
		Trigger_Nschet:         zap.DN_GIS.Trigger_Nschet,
		Trigger_Dschet:         parseDate(zap.DN_GIS.Trigger_Dschet),
		Trigger_Idcase:         zap.DN_GIS.Trigger_Idcase,
		Trigger_SL_Id:          zap.DN_GIS.Trigger_SL_Id,
		Trigger_SL_Nhistory:    zap.DN_GIS.Trigger_SL_Nhistory,
		Trigger_DS_CD:          zap.DN_GIS.Trigger_DS_CD,
		Trigger_MCode:          zap.DN_GIS.Trigger_MCode,
		Trigger_DT:             parseDate(zap.DN_GIS.Trigger_DT),

		// DN_LIST
		DN_List_Period_CD:     zap.DN_LIST.DN_List_Period_CD,
		DN_List_Filename:      zap.DN_LIST.DN_List_Filename,
		CODE_L:                zap.DN_LIST.CODE_L,
		DN_List_Result_Code:   zap.DN_LIST.DN_List_Result_Code,
		DN_List_Date_Checking: parseDate(zap.DN_LIST.DN_List_Date_Checking),
		DN_List_Result_Descr:  zap.DN_LIST.DN_List_Result_Descr,

		// DN_PLAN
		DN_Plan_Period:        parseDate(zap.DN_PLAN.DN_Plan_Period),
		DN_Plan_Filename:      zap.DN_PLAN.DN_Plan_Filename,
		CODE_P:                zap.DN_PLAN.CODE_P,
		DN_Plan_Result_Code:   zap.DN_PLAN.DN_Plan_Result_Code,
		DN_Plan_Date_Checking: parseDate(zap.DN_PLAN.DN_Plan_Date_Checking),
		DN_Plan_Result_Descr:  zap.DN_PLAN.DN_Plan_Result_Descr,

		// Служебные
		Insert_DTTM: parseDateTime(zap.Insert_DTTM, now),
		Update_DTTM: parseDateTime(zap.Update_DTTM, now),
	}
}

// Сохранение батча с использованием транзакции
func (p *GFParser) saveBatch(records []GFRecord) error {
	if len(records) == 0 {
		return nil
	}

	// Начинаем транзакцию
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback()

	// Подготавливаем запрос
	stmt, err := tx.Prepare(`
        INSERT INTO GF_VerificationResults (
            Version, FileType, DataDate, FileName, RegionCode, Period, RecordsCount,
            DN_Patient_Id, ENP, GenderCode, BirthDate, SMO_Code, Attach_MCode, 
            Attach_Date, SMO_Region_Code, Group_RH_Code, Group_RH_DS, DN_Prvs,
            Group_RH_Profile, Group_RH_Name, DN_Rule_In_Name,
            Trigger_Schet_Filename, Trigger_Schet_Code, Trigger_Nschet,
            Trigger_Dschet, Trigger_Idcase, Trigger_SL_Id, Trigger_SL_Nhistory,
            Trigger_DS_CD, Trigger_MCode, Trigger_DT,
            DN_List_Period_CD, DN_List_Filename, CODE_L, DN_List_Result_Code,
            DN_List_Date_Checking, DN_List_Result_Descr,
            DN_Plan_Period, DN_Plan_Filename, CODE_P, DN_Plan_Result_Code,
            DN_Plan_Date_Checking, DN_Plan_Result_Descr,
            Insert_DTTM, Update_DTTM
        ) VALUES (
            @Version, @FileType, @DataDate, @FileName, @RegionCode, @Period, @RecordsCount,
            @DN_Patient_Id, @ENP, @GenderCode, @BirthDate, @SMO_Code, @Attach_MCode,
            @Attach_Date, @SMO_Region_Code, @Group_RH_Code, @Group_RH_DS, @DN_Prvs,
            @Group_RH_Profile, @Group_RH_Name, @DN_Rule_In_Name,
            @Trigger_Schet_Filename, @Trigger_Schet_Code, @Trigger_Nschet,
            @Trigger_Dschet, @Trigger_Idcase, @Trigger_SL_Id, @Trigger_SL_Nhistory,
            @Trigger_DS_CD, @Trigger_MCode, @Trigger_DT,
            @DN_List_Period_CD, @DN_List_Filename, @CODE_L, @DN_List_Result_Code,
            @DN_List_Date_Checking, @DN_List_Result_Descr,
            @DN_Plan_Period, @DN_Plan_Filename, @CODE_P, @DN_Plan_Result_Code,
            @DN_Plan_Date_Checking, @DN_Plan_Result_Descr,
            @Insert_DTTM, @Update_DTTM
        )`)
	if err != nil {
		return fmt.Errorf("ошибка подготовки запроса: %w", err)
	}
	defer stmt.Close()

	// Выполняем вставку для каждой записи в батче
	for _, record := range records {
		_, err := stmt.Exec(
			sql.Named("Version", record.Version),
			sql.Named("FileType", record.FileType),
			sql.Named("DataDate", record.DataDate),
			sql.Named("FileName", record.FileName),
			sql.Named("RegionCode", record.RegionCode),
			sql.Named("Period", record.Period),
			sql.Named("RecordsCount", record.RecordsCount),

			sql.Named("DN_Patient_Id", record.DN_Patient_ID),
			sql.Named("ENP", record.ENP),
			sql.Named("GenderCode", record.GenderCode),
			sql.Named("BirthDate", record.BirthDate),
			sql.Named("SMO_Code", record.SMO_Code),
			sql.Named("Attach_MCode", record.Attach_MCode),
			sql.Named("Attach_Date", record.Attach_Date),
			sql.Named("SMO_Region_Code", record.SMO_Region_Code),
			sql.Named("Group_RH_Code", record.Group_RH_Code),
			sql.Named("Group_RH_DS", record.Group_RH_DS),
			sql.Named("DN_Prvs", record.DN_Prvs),
			sql.Named("Group_RH_Profile", record.Group_RH_Profile),
			sql.Named("Group_RH_Name", record.Group_RH_Name),
			sql.Named("DN_Rule_In_Name", record.DN_Rule_In_Name),

			sql.Named("Trigger_Schet_Filename", record.Trigger_Schet_Filename),
			sql.Named("Trigger_Schet_Code", record.Trigger_Schet_Code),
			sql.Named("Trigger_Nschet", record.Trigger_Nschet),
			sql.Named("Trigger_Dschet", record.Trigger_Dschet),
			sql.Named("Trigger_Idcase", record.Trigger_Idcase),
			sql.Named("Trigger_SL_Id", record.Trigger_SL_Id),
			sql.Named("Trigger_SL_Nhistory", record.Trigger_SL_Nhistory),
			sql.Named("Trigger_DS_CD", record.Trigger_DS_CD),
			sql.Named("Trigger_MCode", record.Trigger_MCode),
			sql.Named("Trigger_DT", record.Trigger_DT),

			sql.Named("DN_List_Period_CD", record.DN_List_Period_CD),
			sql.Named("DN_List_Filename", record.DN_List_Filename),
			sql.Named("CODE_L", record.CODE_L),
			sql.Named("DN_List_Result_Code", record.DN_List_Result_Code),
			sql.Named("DN_List_Date_Checking", record.DN_List_Date_Checking),
			sql.Named("DN_List_Result_Descr", record.DN_List_Result_Descr),

			sql.Named("DN_Plan_Period", record.DN_Plan_Period),
			sql.Named("DN_Plan_Filename", record.DN_Plan_Filename),
			sql.Named("CODE_P", record.CODE_P),
			sql.Named("DN_Plan_Result_Code", record.DN_Plan_Result_Code),
			sql.Named("DN_Plan_Date_Checking", record.DN_Plan_Date_Checking),
			sql.Named("DN_Plan_Result_Descr", record.DN_Plan_Result_Descr),

			sql.Named("Insert_DTTM", record.Insert_DTTM),
			sql.Named("Update_DTTM", record.Update_DTTM),
		)

		if err != nil {
			return fmt.Errorf("ошибка вставки записи (Patient_ID: %s): %w",
				record.DN_Patient_ID, err)
		}
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return nil
}

// Вспомогательные функции парсинга дат
func parseDate(dateStr string) interface{} {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return nil
	}

	formats := []string{"2006-01-02", "2006-01-02 15:04:05", "2006-01-02T15:04:05"}
	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t.Format("2006-01-02")
		}
	}
	return nil
}

func parseDateTime(dateStr string, defaultVal string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return defaultVal
	}

	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000",
		"2006-01-02 15:04:05.000",
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
	}
	return defaultVal
}

// Обработка директории с XML файлами
func (p *GFParser) ParseDirectory(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	log.Printf("Найдено файлов в директории: %d", len(files))

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".xml") {
			continue
		}

		filePath := dirPath + "/" + file.Name()
		fileSize, _ := os.Stat(filePath)
		log.Printf("Начало обработки файла: %s (размер: %d байт)", filePath, fileSize.Size())

		startTime := time.Now()
		if err := p.ParseFile(filePath); err != nil {
			log.Printf("Ошибка обработки файла %s: %v", filePath, err)
			continue
		}

		log.Printf("Файл %s обработан за %v", file.Name(), time.Since(startTime))
	}

	log.Printf("Итог: Всего сохранено записей: %d, Ошибок: %d", p.totalSaved, p.totalErrors)
	return nil
}

// Основная функция
func main() {
	if len(os.Args) < 2 {
		log.Fatal("Использование: go run gf.go <путь_к_xml_файлу>")
	}

	filePath := os.Args[1]

	// Загружаем переменные из .env (по умолчанию ищет файл .env в текущей директории)
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}

	sqlPort, _ := strconv.Atoi(os.Getenv("SQL_PORT"))

	config := DBConfig{
		Server:   os.Getenv("SQL_SERVER_NAME"),
		Port:     sqlPort,
		User:     os.Getenv("SQL_USER"),
		Password: os.Getenv("SQL_PASSWORD"),
		Database: os.Getenv("SQL_DATABASE"),
	}

	parser, err := NewGFParser(config)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer parser.Close()

	// Парсинг одного файла
	err = parser.ParseFile(filePath)

	// Парсинг всех XML файлов в директории
	// err = parser.ParseDirectory("./xml_files")
	if err != nil {
		log.Fatal("Ошибка обработки:", err)
	}

	log.Println("Обработка завершена успешно")
}
