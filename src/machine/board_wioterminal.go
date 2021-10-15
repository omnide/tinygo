// +build wioterminal

package machine

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xf01669ef

const (
	ADC0 = A0
	ADC1 = A1
	ADC2 = A2
	ADC3 = A3
	ADC4 = A4
	ADC5 = A5
	ADC6 = A6
	ADC7 = A7
	ADC8 = A8

	LED    = PIN_LED
	BUTTON = BUTTON_1
)

const (
	// https://github.com/Seeed-Studio/ArduinoCore-samd/blob/master/variants/wio_terminal/variant.h

	// LEDs
	PIN_LED_13   = PA15
	PIN_LED_RXL  = PA15
	PIN_LED_TXL  = PA15
	PIN_LED      = PIN_LED_13
	PIN_LED2     = PIN_LED_RXL
	PIN_LED3     = PIN_LED_TXL
	LED_BUILTIN  = PIN_LED_13
	PIN_NEOPIXEL = PA15

	//Digital PINs
	D0 = PB08
	D1 = PB09
	D2 = PA07
	D3 = PB04
	D4 = PB05
	D5 = PB06
	D6 = PA04
	D7 = PB07
	D8 = PA06

	//Analog PINs
	A0 = PB08 // ADC/AIN[0]
	A1 = PB09 // ADC/AIN[2]
	A2 = PA07 // ADC/AIN[3]
	A3 = PB04 // ADC/AIN[4]
	A4 = PB05 // ADC/AIN[5]
	A5 = PB06 // ADC/AIN[10]
	A6 = PA04 // ADC/AIN[10]
	A7 = PB07 // ADC/AIN[10]
	A8 = PA06 // ADC/AIN[10]

	// 3.3V    ||    5V
	// BCM2    ||    5V
	// BCM3    ||    GND
	// BCM4    ||    BCM14
	// GND     ||    BCM15
	// BCM17   ||    BCM18
	// BCM27   ||    GND
	// BCM22   ||    BCM23
	// GND     ||    BCM24
	// BCM10   ||    GND
	// BCM9    ||    BCM25
	// BCM11   ||    BCM8
	// GND     ||    BCM7
	// BCM0    ||    BCM1
	// BCM5    ||    GND
	// BCM6    ||    BCM12
	// BCM13   ||    GND
	// BCM19   ||    BCM16
	// BCM26   ||    BCM20
	// GND     ||    BCM21

	//PIN DEFINE FOR RPI
	BCM0  = PA13 // I2C Wire1
	BCM1  = PA12 // I2C Wire1
	BCM2  = PA17 // I2C Wire2
	BCM3  = PA16 // I2C Wire2
	BCM4  = PB14 // GCLK
	BCM5  = PB12 // GCLK
	BCM6  = PB13 // GCLK
	BCM7  = PA05 // DAC1
	BCM8  = PB01 // SPI SS
	BCM9  = PB00 // SPI SDI
	BCM10 = PB02 // SPI SDO
	BCM11 = PB03 // SPI SCK
	BCM12 = PB06
	BCM13 = PA07
	BCM14 = PB27 // UART Serial1
	BCM15 = PB26 // UART Serial1
	BCM16 = PB07
	BCM17 = PA02 // DAC0
	BCM18 = PB28 // FPC Digital & AD pins
	BCM19 = PA20 // WIO_IR
	BCM20 = PA21 // I2S SDO
	BCM21 = PA22 // I2S SDI
	BCM22 = PB09
	BCM23 = PA07
	BCM24 = PB04
	BCM25 = PB05
	BCM26 = PA06
	BCM27 = PB08

	// FPC NEW DEFINE
	FPC1  = PB28 // FPC Digital & AD pins
	FPC2  = PB17
	FPC3  = PB29
	FPC4  = PA14
	FPC5  = PC01
	FPC6  = PC02
	FPC7  = PC03
	FPC8  = PC04
	FPC9  = PC31
	FPC10 = PD00

	// RPI Analog RPIs
	RPI_A0 = PB08
	RPI_A1 = PB09
	RPI_A2 = PA07
	RPI_A3 = PB04
	RPI_A4 = PB05
	RPI_A5 = PB06
	RPI_A6 = PA04
	RPI_A7 = PB07
	RPI_A8 = PA06

	PIN_DAC0 = PA02
	PIN_DAC1 = PA05

	// FPO Analog RPIs
	//FPC_A7  = FPC_D7
	//FPC_A8  = FPC_D8
	//FPC_A9  = FPC_D9
	//FPC_A11 = FPC_D11
	//FPC_A12 = FPC_D12
	//FPC_A13 = FPC_D13

	// USB
	PIN_USB_DM          = PA24
	PIN_USB_DP          = PA25
	PIN_USB_HOST_ENABLE = PA27

	// BUTTON
	BUTTON_1  = PC26
	BUTTON_2  = PC27
	BUTTON_3  = PC28
	WIO_KEY_A = PC26
	WIO_KEY_B = PC27
	WIO_KEY_C = PC28

	// SWITCH
	SWITCH_X = PD20
	SWITCH_Y = PD12
	SWITCH_Z = PD09
	SWITCH_B = PD08
	SWITCH_U = PD10

	WIO_5S_UP    = PD20
	WIO_5S_LEFT  = PD12
	WIO_5S_RIGHT = PD09
	WIO_5S_DOWN  = PD08
	WIO_5S_PRESS = PD10

	// IRQ0 : RTL8720D
	IRQ0 = PC20

	// BUZZER_CTR
	BUZZER_CTR = PD11
	WIO_BUZZER = PD11

	// MIC_INPUT
	MIC_INPUT = PC30
	WIO_MIC   = PC30

	// GCLK
	GCLK0 = PB14
	GCLK1 = PB12
	GCLK2 = PB13

	// Serial interfaces
	// Serial1
	PIN_SERIAL1_RX = PB27
	PIN_SERIAL1_TX = PB26

	// Serial2 : RTL8720D
	PIN_SERIAL2_RX = PC23
	PIN_SERIAL2_TX = PC22

	// Wire Interfaces
	// I2C Wire2
	// I2C1
	PIN_WIRE_SDA = PA17
	PIN_WIRE_SCL = PA16
	SDA          = PIN_WIRE_SDA
	SCL          = PIN_WIRE_SCL

	// I2C Wire1
	// I2C0 : LIS3DHTR and ATECC608
	PIN_WIRE1_SDA = PA13
	PIN_WIRE1_SCL = PA12

	SDA1 = PIN_WIRE1_SDA
	SCL1 = PIN_WIRE1_SCL

	PIN_GYROSCOPE_WIRE_SDA = PIN_WIRE1_SDA
	PIN_GYROSCOPE_WIRE_SCL = PIN_WIRE1_SCL
	GYROSCOPE_INT1         = PC21

	WIO_LIS3DH_SDA = PIN_WIRE1_SDA
	WIO_LIS3DH_SCL = PIN_WIRE1_SCL
	WIO_LIS3DH_INT = PC21

	// SPI
	PIN_SPI_SDI = PB00
	PIN_SPI_SDO = PB02
	PIN_SPI_SCK = PB03
	PIN_SPI_SS  = PB01

	SS  = PIN_SPI_SS
	SDO = PIN_SPI_SDO
	SDI = PIN_SPI_SDI
	SCK = PIN_SPI_SCK

	// SPI1 RTL8720D_SPI
	PIN_SPI1_SDI = PC24
	PIN_SPI1_SDO = PB24
	PIN_SPI1_SCK = PB25
	PIN_SPI1_SS  = PC25

	SS1  = PIN_SPI1_SS
	SDO1 = PIN_SPI1_SDO
	SDI1 = PIN_SPI1_SDI
	SCK1 = PIN_SPI1_SCK

	// SPI2 SD_SPI
	PIN_SPI2_SDI = PC18
	PIN_SPI2_SDO = PC16
	PIN_SPI2_SCK = PC17
	PIN_SPI2_SS  = PC19

	SS2  = PIN_SPI2_SS
	SDO2 = PIN_SPI2_SDO
	SDI2 = PIN_SPI2_SDI
	SCK2 = PIN_SPI2_SCK

	// SPI3 LCD_SPI
	PIN_SPI3_SDI = PB18
	PIN_SPI3_SDO = PB19
	PIN_SPI3_SCK = PB20
	PIN_SPI3_SS  = PB21

	SS3  = PIN_SPI3_SS
	SDO3 = PIN_SPI3_SDO
	SDI3 = PIN_SPI3_SDI
	SCK3 = PIN_SPI3_SCK

	// Needed for SD library
	SDCARD_SDI_PIN = PIN_SPI2_SDI
	SDCARD_SDO_PIN = PIN_SPI2_SDO
	SDCARD_SCK_PIN = PIN_SPI2_SCK
	SDCARD_SS_PIN  = PIN_SPI2_SS
	SDCARD_DET_PIN = PD21

	LCD_SDI_PIN   = PIN_SPI3_SDI
	LCD_SDO_PIN   = PIN_SPI3_SDO
	LCD_SCK_PIN   = PIN_SPI3_SCK
	LCD_SS_PIN    = PIN_SPI3_SS
	LCD_DC        = PC06
	LCD_RESET     = PC07
	LCD_BACKLIGHT = PC05

	// 4 WIRE LCD TOUCH
	LCD_XL = PC10
	LCD_YU = PC11
	LCD_XR = PC12
	LCD_YD = PC13

	// Needed for RTL8720D
	RTL8720D_SDI_PIN = PIN_SPI1_SDI
	RTL8720D_SDO_PIN = PIN_SPI1_SDO
	RTL8720D_SCK_PIN = PIN_SPI1_SCK
	RTL8720D_SS_PIN  = PIN_SPI1_SS

	//QSPI Pins
	PIN_QSPI_IO0 = PA08
	PIN_QSPI_IO1 = PA09
	PIN_QSPI_IO2 = PA10
	PIN_QSPI_IO3 = PA11
	PIN_QSPI_SCK = PB10
	PIN_QSPI_CS  = PB11

	// I2S Interfaces
	PIN_I2S_FS  = PA20
	PIN_I2S_SCK = PB16
	PIN_I2S_SDO = PA22
	PIN_I2S_SDI = PA21

	I2S_LRCLK = PA20
	I2S_BLCK  = PB16
	I2S_SDOUT = PA22
	I2S_SDIN  = PA21

	// RTL8720D Interfaces
	RTL8720D_CHIP_PU = PA18
	RTL8720D_GPIO0   = PA19 // SYNC

	// SWD
	SWDCLK = PA30
	SWDIO  = PA31
	SWO    = PB30

	// light sensor
	WIO_LIGHT = PD01

	// ir sensor
	WIO_IR = PB31

	// OUTPUT_CTR
	OUTPUT_CTR_5V  = PC14
	OUTPUT_CTR_3V3 = PC15
)

// USBCDC pins
const (
	USBCDC_DM_PIN = PIN_USB_DM
	USBCDC_DP_PIN = PIN_USB_DP
)

// UART1 pins
const (
	UART_TX_PIN = PIN_SERIAL1_TX
	UART_RX_PIN = PIN_SERIAL1_RX
)

// UART2 pins RTL8720D
const (
	UART2_TX_PIN = PIN_SERIAL2_TX
	UART2_RX_PIN = PIN_SERIAL2_RX
)

var (
	UART1 = &sercomUSART2

	// RTL8720D
	UART2 = &sercomUSART1
)

// I2C pins
const (
	SDA0_PIN = PIN_WIRE_SDA // SDA: SERCOM3/PAD[0]
	SCL0_PIN = PIN_WIRE_SCL // SCL: SERCOM3/PAD[1]

	SDA1_PIN = PIN_WIRE1_SDA // SDA: SERCOM4/PAD[0]
	SCL1_PIN = PIN_WIRE1_SCL // SCL: SERCOM4/PAD[1]

	SDA_PIN = SDA0_PIN
	SCL_PIN = SCL0_PIN
)

// SPI pins
const (
	SPI0_SCK_PIN = SCK // SCK:  SERCOM5/PAD[1]
	SPI0_SDO_PIN = SDO // SDO: SERCOM5/PAD[0]
	SPI0_SDI_PIN = SDI // SDI: SERCOM5/PAD[2]

	// RTL8720D
	SPI1_SCK_PIN = SCK1 // SCK:  SERCOM0/PAD[1]
	SPI1_SDO_PIN = SDO1 // SDO: SERCOM0/PAD[0]
	SPI1_SDI_PIN = SDI1 // SDI: SERCOM0/PAD[2]

	// SD
	SPI2_SCK_PIN = SCK2 // SCK:  SERCOM6/PAD[1]
	SPI2_SDO_PIN = SDO2 // SDO: SERCOM6/PAD[0]
	SPI2_SDI_PIN = SDI2 // SDI: SERCOM6/PAD[2]

	// LCD
	SPI3_SCK_PIN = SCK3 // SCK:  SERCOM7/PAD[1]
	SPI3_SDO_PIN = SDO3 // SDO: SERCOM7/PAD[3]
	SPI3_SDI_PIN = SDI3 // SDI: SERCOM7/PAD[2]
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Seeed Wio Terminal"
	usb_STRING_MANUFACTURER = "Seeed"
)

var (
	usb_VID uint16 = 0x2886
	usb_PID uint16 = 0x802D
)
