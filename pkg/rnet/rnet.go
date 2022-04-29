package rnet

const (
	MAX_XY_DATA int8    = 100
	MIN_XY_DATA int8    = -100
	MAX_XY_JOY  float32 = 1.0
	MIN_XY_JOY  float32 = -1.0
	LIMIT_X_POS int8    = 40  // derived from testing, the joy forward seems to stop at 0x28 which is 0d40
	LIMIT_X_NEG int8    = -40 // assume backwards is the same
	LIMIT_Y_POS int8    = 100 // sideways joystick goes to +/- 100
	LIMIT_Y_NEG int8    = -100
	// THRESHOLD_SQ is the squared min threshold, used to create a circle of dead zone around the joystick in a zero position.
	THRESHOLD_SQ float32 = 0.01 // TODO, determine the proper value with testing.
)

func ConvertJoyToData(joyx, joyy float32) (x, y int8) {
	// dead zone in center position
	if joyx*joyx <= THRESHOLD_SQ {
		joyx = 0.0
	}
	if joyy*joyy <= THRESHOLD_SQ {
		joyy = 0.0
	}

	// clamp values to min and max accepted inputs (-1.0 to 1.0)
	if joyx > MAX_XY_JOY {
		joyx = MAX_XY_JOY
	}
	if joyx < MIN_XY_JOY {
		joyx = MIN_XY_JOY
	}
	if joyy > MAX_XY_JOY {
		joyy = MAX_XY_JOY
	}
	if joyy < MIN_XY_JOY {
		joyy = MIN_XY_JOY
	}

	// convert input to value within output range (-100 to 100)
	factor := (float32(MAX_XY_DATA) - float32(MIN_XY_DATA)) / (MAX_XY_JOY - MIN_XY_JOY)
	xx, yy := int8(joyx*factor), int8(joyy*factor)
	if xx > LIMIT_X_POS {
		xx = LIMIT_X_POS
	}
	if xx < LIMIT_X_NEG {
		xx = LIMIT_X_NEG
	}
	if yy > LIMIT_Y_POS {
		yy = LIMIT_Y_POS
	}
	if yy < LIMIT_Y_NEG {
		yy = LIMIT_Y_NEG
	}
	return xx, yy
}

const (
	ARB_ID_MASK uint32 = 0x82000F00 // note this includes the CAN_EFF_FLAG 0x80000000
	JSM_ID_MASK uint32 = 0x00000F00
)

func IsMovementFrame(id uint32) bool {
	// the JSM_ID nibble is "don't care", so we set it to F, then compare to the ARB_ID mask
	// e.g.: 0x82000F00^0x82000F00 --> 0x00000000 == 0 --> true
	//   and 0x12345F78^0x82000F00 --> 0x90345078 == 0 --> false
	return (id|JSM_ID_MASK)^ARB_ID_MASK == 0
}

func GetJID(id uint32) uint8 {
	// we just mask off the JSM ID nibble and shift it right
	// e.g. "5": 0x82000500 -mask-> 0x00000500 -shift-> 0x5
	return uint8((id & JSM_ID_MASK) >> 8)
}

const (
	INPUT_SCALE_SIDE = -1.0
	INPUT_SCALE_FWD  = 1.0
)

func ConvertDataToJoy(xx, yy uint8) (fwd, side float64) {
	x, y := int8(xx), int8(yy)
	if x != 0 {
		side = INPUT_SCALE_SIDE * float64(x) / 100.0
	}
	if y != 0 {
		fwd = INPUT_SCALE_FWD * float64(y) / 100.0
	}
	return fwd, side
}

// // we assume a signed hex input between -100 (0x9C) and 100 (0x64)
// // note: +100 = 0x64 and -100 = 0x9C (two's complement of 0x64)
// func GetXY(id uint32) (x, y int8) {
// 	if !IsMovementFrame(id) {
// 		return 0, 0
// 	}

// }

// func ConvertDataToJoy(x, y int8) (joyx, joyy float32) {
// 	// dead zone in center position
// 	if joyx*joyx <= THRESHOLD_SQ {
// 		joyx = 0.0
// 	}
// 	if joyy*joyy <= THRESHOLD_SQ {
// 		joyy = 0.0
// 	}

// 	// clamp values to min and max accepted inputs (-1.0 to 1.0)
// 	if joyx > MAX_XY_JOY {
// 		joyx = MAX_XY_JOY
// 	}
// 	if joyx < MIN_XY_JOY {
// 		joyx = MIN_XY_JOY
// 	}
// 	if joyy > MAX_XY_JOY {
// 		joyy = MAX_XY_JOY
// 	}
// 	if joyy < MIN_XY_JOY {
// 		joyy = MIN_XY_JOY
// 	}

// 	// convert input to value within output range (-100 to 100)
// 	factor := (float32(MAX_XY_DATA) - float32(MIN_XY_DATA)) / (MAX_XY_JOY - MIN_XY_JOY)
// 	return int8(joyx * factor), int8(joyy * factor)
// }
