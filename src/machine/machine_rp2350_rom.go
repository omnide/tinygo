//go:build tinygo && rp2350

package machine

import ()

/*
typedef unsigned char uint8_t;
typedef unsigned short uint16_t;
typedef unsigned long uint32_t;
typedef unsigned long size_t;
typedef unsigned long uintptr_t;

#define false 0
#define true 1
typedef int bool;

// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_platform_compiler/include/pico/platform/compiler.h

#define pico_default_asm_volatile(...) __asm volatile (".syntax unified\n" __VA_ARGS__)


// https://github.com/raspberrypi/pico-sdk
// src/rp2350/pico_platform/include/pico/platform.h

static bool pico_processor_state_is_nonsecure(void) {
//    // todo add a define to disable NS checking at all?
//    // IDAU-Exempt addresses return S=1 when tested in the Secure state,
//    // whereas executing a tt in the NonSecure state will always return S=0.
//    uint32_t tt;
//    pico_default_asm_volatile (
//        "movs %0, #0\n"
//        "tt %0, %0\n"
//        : "=r" (tt) : : "cc"
//    );
//    return !(tt & (1u << 22));

    return false;
}


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_bootrom/include/pico/bootrom_constants.h

// RP2040 & RP2350
#define ROM_DATA_SOFTWARE_GIT_REVISION          ROM_TABLE_CODE('G', 'R')
#define ROM_FUNC_FLASH_ENTER_CMD_XIP            ROM_TABLE_CODE('C', 'X')
#define ROM_FUNC_FLASH_EXIT_XIP                 ROM_TABLE_CODE('E', 'X')
#define ROM_FUNC_FLASH_FLUSH_CACHE              ROM_TABLE_CODE('F', 'C')
#define ROM_FUNC_CONNECT_INTERNAL_FLASH         ROM_TABLE_CODE('I', 'F')
#define ROM_FUNC_FLASH_RANGE_ERASE              ROM_TABLE_CODE('R', 'E')
#define ROM_FUNC_FLASH_RANGE_PROGRAM            ROM_TABLE_CODE('R', 'P')

// RP2350 only
#define ROM_FUNC_PICK_AB_PARTITION              ROM_TABLE_CODE('A', 'B')
#define ROM_FUNC_CHAIN_IMAGE                    ROM_TABLE_CODE('C', 'I')
#define ROM_FUNC_EXPLICIT_BUY                   ROM_TABLE_CODE('E', 'B')
#define ROM_FUNC_FLASH_RUNTIME_TO_STORAGE_ADDR  ROM_TABLE_CODE('F', 'A')
#define ROM_DATA_FLASH_DEVINFO16_PTR            ROM_TABLE_CODE('F', 'D')
#define ROM_FUNC_FLASH_OP                       ROM_TABLE_CODE('F', 'O')
#define ROM_FUNC_GET_B_PARTITION                ROM_TABLE_CODE('G', 'B')
#define ROM_FUNC_GET_PARTITION_TABLE_INFO       ROM_TABLE_CODE('G', 'P')
#define ROM_FUNC_GET_SYS_INFO                   ROM_TABLE_CODE('G', 'S')
#define ROM_FUNC_GET_UF2_TARGET_PARTITION       ROM_TABLE_CODE('G', 'U')
#define ROM_FUNC_LOAD_PARTITION_TABLE           ROM_TABLE_CODE('L', 'P')
#define ROM_FUNC_OTP_ACCESS                     ROM_TABLE_CODE('O', 'A')
#define ROM_DATA_PARTITION_TABLE_PTR            ROM_TABLE_CODE('P', 'T')
#define ROM_FUNC_FLASH_RESET_ADDRESS_TRANS      ROM_TABLE_CODE('R', 'A')
#define ROM_FUNC_REBOOT                         ROM_TABLE_CODE('R', 'B')
#define ROM_FUNC_SET_ROM_CALLBACK               ROM_TABLE_CODE('R', 'C')
#define ROM_FUNC_SECURE_CALL                    ROM_TABLE_CODE('S', 'C')
#define ROM_FUNC_SET_NS_API_PERMISSION          ROM_TABLE_CODE('S', 'P')
#define ROM_FUNC_BOOTROM_STATE_RESET            ROM_TABLE_CODE('S', 'R')
#define ROM_FUNC_SET_BOOTROM_STACK              ROM_TABLE_CODE('S', 'S')
#define ROM_DATA_SAVED_XIP_SETUP_FUNC_PTR       ROM_TABLE_CODE('X', 'F')
#define ROM_FUNC_FLASH_SELECT_XIP_READ_MODE     ROM_TABLE_CODE('X', 'M')
#define ROM_FUNC_VALIDATE_NS_BUFFER             ROM_TABLE_CODE('V', 'B')

#define BOOTSEL_FLAG_GPIO_PIN_SPECIFIED         0x20

#define BOOTROM_FUNC_TABLE_OFFSET 0x14

// todo remove this (or #ifdef it for A1/A2)
#define BOOTROM_IS_A2() ((*(volatile uint8_t *)0x13) == 2)
#define BOOTROM_WELL_KNOWN_PTR_SIZE (BOOTROM_IS_A2() ? 2 : 4)

#define BOOTROM_VTABLE_OFFSET 0x00
#define BOOTROM_TABLE_LOOKUP_OFFSET     (BOOTROM_FUNC_TABLE_OFFSET + BOOTROM_WELL_KNOWN_PTR_SIZE)

// https://github.com/raspberrypi/pico-sdk
// src/common/boot_picoboot_headers/include/boot/picoboot_constants.h

// values 0-7 are secure/non-secure
#define REBOOT2_FLAG_REBOOT_TYPE_NORMAL       0x0 // param0 = diagnostic partition
#define REBOOT2_FLAG_REBOOT_TYPE_BOOTSEL      0x2 // param0 = bootsel_flags, param1 = gpio_config
#define REBOOT2_FLAG_REBOOT_TYPE_RAM_IMAGE    0x3 // param0 = image_base, param1 = image_end
#define REBOOT2_FLAG_REBOOT_TYPE_FLASH_UPDATE 0x4 // param0 = update_base

#define REBOOT2_FLAG_NO_RETURN_ON_SUCCESS    0x100

#define RT_FLAG_FUNC_ARM_SEC    0x0004
#define RT_FLAG_FUNC_ARM_NONSEC 0x0010

// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_bootrom/include/pico/bootrom.h

#define ROM_TABLE_CODE(c1, c2) ((c1) | ((c2) << 8))

typedef void *(*rom_table_lookup_fn)(uint32_t code, uint32_t mask);

__attribute__((always_inline))
static void *rom_func_lookup_inline(uint32_t code) {
    rom_table_lookup_fn rom_table_lookup = (rom_table_lookup_fn) (uintptr_t)*(uint16_t*)(BOOTROM_TABLE_LOOKUP_OFFSET);
    if (pico_processor_state_is_nonsecure()) {
        return rom_table_lookup(code, RT_FLAG_FUNC_ARM_NONSEC);
    } else {
        return rom_table_lookup(code, RT_FLAG_FUNC_ARM_SEC);
    }
}


typedef int (*rom_reboot_fn)(uint32_t flags, uint32_t delay_ms, uint32_t p0, uint32_t p1);

__attribute__((always_inline))
int rom_reboot(uint32_t flags, uint32_t delay_ms, uint32_t p0, uint32_t p1) {
    rom_reboot_fn func = (rom_reboot_fn) rom_func_lookup_inline(ROM_FUNC_REBOOT);
    return func(flags, delay_ms, p0, p1);
}


// https://github.com/raspberrypi/pico-sdk
// src/rp2_common/pico_bootrom/bootrom.c


void reset_usb_boot(uint32_t usb_activity_gpio_pin_mask, uint32_t disable_interface_mask) {
    uint32_t flags = disable_interface_mask;
    if (usb_activity_gpio_pin_mask) {
        flags |= BOOTSEL_FLAG_GPIO_PIN_SPECIFIED;
        // the parameter is actually the gpio number, but we only care if BOOTSEL_FLAG_GPIO_PIN_SPECIFIED
        usb_activity_gpio_pin_mask = (uint32_t)__builtin_ctz(usb_activity_gpio_pin_mask);
    }
    rom_reboot(REBOOT2_FLAG_REBOOT_TYPE_BOOTSEL | REBOOT2_FLAG_NO_RETURN_ON_SUCCESS, 10, flags, usb_activity_gpio_pin_mask);
    __builtin_unreachable();
}


*/
import "C"

func enterBootloader() {
	C.reset_usb_boot(0, 0)
}
