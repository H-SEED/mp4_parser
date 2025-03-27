package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Box 구조체: MP4 박스의 기본 구조
type Box struct {
	Size uint32
	Type [4]byte
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("사용법: go run test.go <mp4 파일 경로>")
		return
	}

	// MP4 파일 열기
	filePath := os.Args[1]
	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println("파일 열기 오류:", err)
		return
	}

	defer file.Close()

	// 파일을 순차적으로 읽으며 박스 파싱
	for {
		// 현재 오프셋 확인
		offset, _ := file.Seek(0, os.SEEK_CUR)

		// 박스 크기와 타입 읽기
		var box Box
		err := binary.Read(file, binary.BigEndian, &box.Size)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Main loop size error:", err)
			break // EOF 또는 오류 발생 시 종료
		}
		err = binary.Read(file, binary.BigEndian, &box.Type)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Main loop type error:", err)
			break
		}

		// 박스 타입을 문자열로 변환
		boxType := string(box.Type[:])
		fmt.Printf("Box: %s, Size: %d, Offset: 0x%X\n", boxType, box.Size, offset)

		// 박스별 처리
		switch boxType {
		case "ftyp":
			parseFtyp(file, box.Size-8)
		case "moov":
			parseMoov(file, box.Size-8)
		default:
			// 기타 박스는 건너뛰기
			_, err := file.Seek(int64(box.Size-8), os.SEEK_CUR)
			if err != nil {
				fmt.Println("Seek error in main loop:", err)
				break
			}
		}
	}
}

// ftyp 박스 파싱
func parseFtyp(file *os.File, remainingSize uint32) {
	/*
	 * ftyp box
	 * major_brand(4 Byte)
	 * minor_version(4 Byte)
	 * compatible_brands(가변 Byte)
	 */
	var majorBrand [4]byte
	var minorVersion uint32
	if err := binary.Read(file, binary.BigEndian, &majorBrand); err != nil {
		fmt.Println("Error reading majorBrand:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &minorVersion); err != nil {
		fmt.Println("Error reading minorVersion:", err)
		return
	}

	fmt.Println("ftyp:")
	fmt.Printf("  major_brand: %s\n", string(majorBrand[:]))
	fmt.Printf("  minor_version: %d\n", minorVersion)

	// compatible_brands는 남은 크기만큼 읽기
	remainingSize -= 8 // major_brand + minor_version
	compatibleBrands := make([]byte, remainingSize)
	if err := binary.Read(file, binary.BigEndian, &compatibleBrands); err != nil {
		fmt.Println("Error reading compatibleBrands:", err)
		return
	}
	for i := 0; i < int(remainingSize); i += 4 {
		fmt.Printf("  compatible_brand: %s\n", string(compatibleBrands[i:i+4]))
	}
}

// moov 박스와 하위 박스 파싱
func parseMoov(file *os.File, remainingSize uint32) {
	parsedSize := uint32(0)

	for parsedSize < remainingSize {
		// 현재 오프셋 확인
		offset, _ := file.Seek(0, os.SEEK_CUR)
		var box Box
		err := binary.Read(file, binary.BigEndian, &box.Size)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println("error :", err)
				break
			}
		}
		err = binary.Read(file, binary.BigEndian, &box.Type)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println("error :", err)
				break
			}
		}

		boxType := string(box.Type[:])
		fmt.Printf("============parse Moov  %s [size:%d, offset:0x%X]============\n", boxType, box.Size, offset)

		// 'mvhd' 또는 'trak' 박스 발견 시 처리
		switch boxType {
		case "mvhd":
			parseMvhd(file, box.Size-8)
		case "trak":
			parseTrak(file, box.Size-8)
		case "tkhd":
			parseTkhd(file, box.Size-8)
		case "edts":
			parseEdts(file, box.Size-8)
		case "mdia":
			parseMdia(file, box.Size-8)
		default:
			_, err := file.Seek(int64(box.Size-8), os.SEEK_CUR)
			if err != nil {
				fmt.Println("Seek error in parseMoov:", err)
				break
			}
		}
		parsedSize += box.Size
	}
}

// mvhd 박스 파싱
func parseMvhd(file *os.File, remainingSize uint32) {
	/* mvhd box
	 * version(1 Byte)
	 * flags(3 Byte)
	 * creation_time(4/8 Byte)
	 * modification_time(4/8 Byte),
	 * timescale(4 Byte)
	 * duration(4/8 Byte)
	 * rate(4 Byte)
	 * volume(2 Byte)
	 * reserved(10 Byte),
	 * matrix(36 Byte)
	 * pre_defined(24 Byte)
	 * next_track_ID(4 Byte)
	 */
	var version uint8
	var flags [3]byte
	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseMvhd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseMvhd:", err)
		return
	}

	var creationTime, modificationTime, timescale, duration uint32
	if version == 1 {
		var ct, mt, dur uint64
		if err := binary.Read(file, binary.BigEndian, &ct); err != nil {
			fmt.Println("Error reading creationTime in parseMvhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &mt); err != nil {
			fmt.Println("Error reading modificationTime in parseMvhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &timescale); err != nil {
			fmt.Println("Error reading timescale in parseMvhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &dur); err != nil {
			fmt.Println("Error reading duration in parseMvhd:", err)
			return
		}
		creationTime = uint32(ct)
		modificationTime = uint32(mt)
		duration = uint32(dur)
	} else {
		if err := binary.Read(file, binary.BigEndian, &creationTime); err != nil {
			fmt.Println("Error reading creationTime in parseMvhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &modificationTime); err != nil {
			fmt.Println("Error reading modificationTime in parseMvhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &timescale); err != nil {
			fmt.Println("Error reading timescale in parseMvhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &duration); err != nil {
			fmt.Println("Error reading duration in parseMvhd:", err)
			return
		}
	}

	// duration(ms) 계산
	durationMs := (float64(duration) * 1000) / float64(timescale)

	// 나머지 데이터는 건너뛰기 (필요 시 추가 파싱 가능)
	_, err := file.Seek(int64(remainingSize-12-(4*uint32(version))), os.SEEK_CUR)
	if err != nil {
		fmt.Println("Seek error in parseMvhd:", err)
		return
	}

	fmt.Println("mvhd:")
	fmt.Printf("  timescale: %d\n", timescale)
	fmt.Printf("  duration: %d\n", duration)
	fmt.Printf("  duration(ms): %.0f\n", durationMs)
}

// trak 박스와 하위 박스 파싱
func parseTrak(file *os.File, remainingSize uint32) {
	parsedSize := uint32(0)

	for parsedSize < remainingSize {
		// 현재 오프셋 확인
		offset, _ := file.Seek(0, os.SEEK_CUR)

		var box Box
		err := binary.Read(file, binary.BigEndian, &box.Size)
		if err != nil {
			fmt.Println("Error reading size in parseTrak:", err)
			break
		}
		err = binary.Read(file, binary.BigEndian, &box.Type)
		if err != nil {
			fmt.Println("Error reading type in parseTrak:", err)
			break
		}

		boxType := string(box.Type[:])
		fmt.Printf("    %s [size:%d, offset:0x%X]\n", boxType, box.Size, offset)

		switch boxType {
		case "tkhd":
			parseTkhd(file, box.Size-8)
		case "edts":
			parseEdts(file, box.Size-8)
		case "mdia":
			parseMdia(file, box.Size-8)
		default:
			_, err := file.Seek(int64(box.Size-8), os.SEEK_CUR)
			if err != nil {
				fmt.Println("Seek error in parseTrak:", err)
				break
			}
		}

		parsedSize += box.Size
	}
}

// tkhd 박스 파싱
func parseTkhd(file *os.File, remainingSize uint32) {
	/* tkhd box
	 * version(1 Byte)
	 * flags(3 Byte)
	 * creation_time(4/8 Byte)
	 * modification_time(4/8 Byte)
	 * track_ID(4 Byte)
	 * reserved(4 Byte)
	 * duration(4/8 Byte)
	 * reserved(8 Byte)
	 * layer(2 Byte)
	 * alternate_group(2 Byte)
	 * volume(2 Byte)
	 * reserved(2 Byte)
	 * matrix(36 Byte)
	 * width(4 Byte)
	 * height(4 Byte)
	 */
	var version uint8
	var flags [3]byte
	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseTkhd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseTkhd:", err)
		return
	}

	// flags에서 enabled 여부 확인 (0x000001 비트가 설정되어 있으면 enabled)
	enabled := (flags[2] & 0x01) == 0x01

	var creationTime, modificationTime, trackID, duration uint32
	if version == 1 {
		var ct, mt, dur uint64
		if err := binary.Read(file, binary.BigEndian, &ct); err != nil {
			fmt.Println("Error reading creationTime in parseTkhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &mt); err != nil {
			fmt.Println("Error reading modificationTime in parseTkhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &trackID); err != nil {
			fmt.Println("Error reading trackID in parseTkhd:", err)
			return
		}
		_, err := file.Seek(4, os.SEEK_CUR) // reserved
		if err != nil {
			fmt.Println("Seek error in parseTkhd (reserved):", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &dur); err != nil {
			fmt.Println("Error reading duration in parseTkhd:", err)
			return
		}
		creationTime = uint32(ct)
		modificationTime = uint32(mt)
		duration = uint32(dur)
	} else {
		if err := binary.Read(file, binary.BigEndian, &creationTime); err != nil {
			fmt.Println("Error reading creationTime in parseTkhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &modificationTime); err != nil {
			fmt.Println("Error reading modificationTime in parseTkhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &trackID); err != nil {
			fmt.Println("Error reading trackID in parseTkhd:", err)
			return
		}
		_, err := file.Seek(4, os.SEEK_CUR) // reserved
		if err != nil {
			fmt.Println("Seek error in parseTkhd (reserved):", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &duration); err != nil {
			fmt.Println("Error reading duration in parseTkhd:", err)
			return
		}
	}

	_, err := file.Seek(8, os.SEEK_CUR) // reserved
	if err != nil {
		fmt.Println("Seek error in parseTkhd (reserved 8 bytes):", err)
		return
	}
	var layer, alternateGroup, volume uint16
	if err := binary.Read(file, binary.BigEndian, &layer); err != nil {
		fmt.Println("Error reading layer in parseTkhd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &alternateGroup); err != nil {
		fmt.Println("Error reading alternateGroup in parseTkhd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &volume); err != nil {
		fmt.Println("Error reading volume in parseTkhd:", err)
		return
	}
	_, err = file.Seek(2, os.SEEK_CUR) // reserved
	if err != nil {
		fmt.Println("Seek error in parseTkhd (reserved 2 bytes):", err)
		return
	}

	// matrix (36바이트)
	var matrix [9]int32
	for i := 0; i < 9; i++ {
		if err := binary.Read(file, binary.BigEndian, &matrix[i]); err != nil {
			fmt.Println("Error reading matrix in parseTkhd:", err)
			return
		}
	}

	// width, height (고정 소수점 16.16 형식)
	var width, height uint32
	if err := binary.Read(file, binary.BigEndian, &width); err != nil {
		fmt.Println("Error reading width in parseTkhd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &height); err != nil {
		fmt.Println("Error reading height in parseTkhd:", err)
		return
	}

	fmt.Println("tkhd:")
	fmt.Printf("  enabled: %d\n", boolToInt(enabled))
	fmt.Printf("  id: %d\n", trackID)
	fmt.Printf("  duration: %d\n", duration)
	fmt.Printf("  volume: %d\n", volume>>8) // 8비트 고정 소수점
	fmt.Printf("  layer: %d\n", layer)
	fmt.Printf("  alternate_group: %d\n", alternateGroup)
	for i := 0; i < 9; i++ {
		fmt.Printf("  matrix_%d: %d, 0x%08X\n", i, matrix[i], uint32(matrix[i]))
	}
	fmt.Printf("  width: %d, 0x%08X\n", width>>16, width)    // 16비트 고정 소수점
	fmt.Printf("  height: %d, 0x%08X\n", height>>16, height) // 16비트 고정 소수점
}

// bool을 int로 변환 (0 또는 1)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// edts 박스와 하위 박스 파싱
func parseEdts(file *os.File, remainingSize uint32) {
	parsedSize := uint32(0)

	for parsedSize < remainingSize {
		// 현재 오프셋 확인
		offset, _ := file.Seek(0, os.SEEK_CUR)

		var box Box
		err := binary.Read(file, binary.BigEndian, &box.Size)
		if err != nil {
			fmt.Println("Error reading size in parseEdts:", err)
			break
		}
		err = binary.Read(file, binary.BigEndian, &box.Type)
		if err != nil {
			fmt.Println("Error reading type in parseEdts:", err)
			break
		}

		boxType := string(box.Type[:])
		fmt.Printf("Edts        %s [size:%d, offset:0x%X]\n", boxType, box.Size, offset)

		if boxType == "elst" {
			parseElst(file, box.Size-8)
		} else {
			_, err := file.Seek(int64(box.Size-8), os.SEEK_CUR)
			if err != nil {
				fmt.Println("Seek error in parseEdts:", err)
				break
			}
		}

		parsedSize += box.Size
	}
}

// elst 박스 파싱
func parseElst(file *os.File, remainingSize uint32) {
	/*
	 * elst box
	 * version(1 Byte)
	 * flags(3 Byte)
	 * entry_count(4 Byte)
	 * entries(가변)
	 */
	var version uint8
	var flags [3]byte
	var entryCount uint32
	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseElst:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseElst:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &entryCount); err != nil {
		fmt.Println("Error reading entryCount in parseElst:", err)
		return
	}

	fmt.Println("elst:")
	fmt.Printf("  entry_count: %d\n", entryCount)

	// 각 엔트리 파싱
	for i := uint32(0); i < entryCount; i++ {
		var segmentDuration, mediaTime uint32
		var mediaRateInt int16
		var mediaRateFrac int16
		if version == 1 {
			var sd, mt uint64
			if err := binary.Read(file, binary.BigEndian, &sd); err != nil {
				fmt.Println("Error reading segmentDuration in parseElst:", err)
				return
			}
			if err := binary.Read(file, binary.BigEndian, &mt); err != nil {
				fmt.Println("Error reading mediaTime in parseElst:", err)
				return
			}
			segmentDuration = uint32(sd)
			mediaTime = uint32(mt)
		} else {
			if err := binary.Read(file, binary.BigEndian, &segmentDuration); err != nil {
				fmt.Println("Error reading segmentDuration in parseElst:", err)
				return
			}
			if err := binary.Read(file, binary.BigEndian, &mediaTime); err != nil {
				fmt.Println("Error reading mediaTime in parseElst:", err)
				return
			}
		}
		if err := binary.Read(file, binary.BigEndian, &mediaRateInt); err != nil {
			fmt.Println("Error reading mediaRateInt in parseElst:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &mediaRateFrac); err != nil {
			fmt.Println("Error reading mediaRateFrac in parseElst:", err)
			return
		}

		fmt.Printf("  entry/segment_duration: %d\n", segmentDuration)
		fmt.Printf("  entry/media_time: %d\n", mediaTime)
		fmt.Printf("  entry/media_rate: %d\n", mediaRateInt)
	}
}

// mdia 박스와 하위 박스 파싱
func parseMdia(file *os.File, remainingSize uint32) {
	parsedSize := uint32(0)

	for parsedSize < remainingSize {
		var box Box
		err := binary.Read(file, binary.BigEndian, &box.Size)
		if err != nil {
			fmt.Println("Error reading size in parseMdia:", err)
			break
		}
		err = binary.Read(file, binary.BigEndian, &box.Type)
		if err != nil {
			fmt.Println("Error reading type in parseMdia:", err)
			break
		}

		boxType := string(box.Type[:])
		switch boxType {
		case "mdhd":
			fmt.Println("parseMdia          Media Information (mdhd) found")
			parseMdhd(file, box.Size-8)
		case "minf":
			fmt.Println("parseMdia          Media Information (minf) found")
			parseMinf(file, box.Size-8)
		default:
			_, err := file.Seek(int64(box.Size-8), os.SEEK_CUR)
			if err != nil {
				fmt.Println("Seek error in parseMdia:", err)
				break
			}
		}

		parsedSize += box.Size
	}

	// 남은 데이터가 있으면 건너뛰기
	if parsedSize < remainingSize {
		_, err := file.Seek(int64(remainingSize-parsedSize), os.SEEK_CUR)
		if err != nil {
			fmt.Println("Seek error in parseMdia (remaining):", err)
		}
	}
}

// mdhd 박스 파싱
func parseMdhd(file *os.File, remainingSize uint32) {
	/*
	 * mdhd box
	 * version(1 Byte)
	 * flags(3 Byte)
	 * creation_time(4/8 Byte)
	 * modification_time(4/8 Byte)
	 * timescale(4 Byte)
	 * duration(4/8 Byte)
	 * language(2 Byte)
	 * pre_defined(2 Byte)
	 */
	var version uint8
	var flags [3]byte
	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseMdhd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseMdhd:", err)
		return
	}

	var creationTime, modificationTime, timescale, duration uint32
	if version == 1 {
		var ct, mt, dur uint64
		if err := binary.Read(file, binary.BigEndian, &ct); err != nil {
			fmt.Println("Error reading creationTime in parseMdhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &mt); err != nil {
			fmt.Println("Error reading modificationTime in parseMdhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &timescale); err != nil {
			fmt.Println("Error reading timescale in parseMdhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &dur); err != nil {
			fmt.Println("Error reading duration in parseMdhd:", err)
			return
		}
		creationTime = uint32(ct)
		modificationTime = uint32(mt)
		duration = uint32(dur)
	} else {
		if err := binary.Read(file, binary.BigEndian, &creationTime); err != nil {
			fmt.Println("Error reading creationTime in parseMdhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &modificationTime); err != nil {
			fmt.Println("Error reading modificationTime in parseMdhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &timescale); err != nil {
			fmt.Println("Error reading timescale in parseMdhd:", err)
			return
		}
		if err := binary.Read(file, binary.BigEndian, &duration); err != nil {
			fmt.Println("Error reading duration in parseMdhd:", err)
			return
		}
	}

	// duration(ms) 계산
	durationMs := (float64(duration) * 1000) / float64(timescale)

	// language (2바이트, ISO-639-2/T 형식)
	var language uint16
	if err := binary.Read(file, binary.BigEndian, &language); err != nil {
		fmt.Println("Error reading language in parseMdhd:", err)
		return
	}
	langStr := parseLanguage(language)

	// pre_defined (2바이트, 건너뛰기)
	_, err := file.Seek(2, os.SEEK_CUR)
	if err != nil {
		fmt.Println("Seek error in parseMdhd (pre_defined):", err)
		return
	}

	fmt.Println("mdhd:")
	fmt.Printf("  timescale: %d\n", timescale)
	fmt.Printf("  duration: %d\n", duration)
	fmt.Printf("  duration(ms): %.0f\n", durationMs)
	fmt.Printf("  language: %s\n", langStr)
}

// language 필드를 ISO-639-2/T 형식으로 변환
func parseLanguage(lang uint16) string {
	// language는 5비트씩 3개의 문자로 나뉨 (15비트 사용)
	// 각 문자는 0x60을 빼고 'a'부터 시작
	char1 := (lang >> 10) & 0x1F
	char2 := (lang >> 5) & 0x1F
	char3 := lang & 0x1F

	// 0x60을 빼고 문자로 변환
	if char1 == 0 && char2 == 0 && char3 == 0 {
		return "und" // 정의되지 않은 언어
	}
	return string([]byte{byte(char1 + 0x60), byte(char2 + 0x60), byte(char3 + 0x60)})
}

// minf 박스와 하위 박스 파싱
func parseMinf(file *os.File, remainingSize uint32) {
	parsedSize := uint32(0)

	for parsedSize < remainingSize {
		// 현재 오프셋 확인
		offset, _ := file.Seek(0, os.SEEK_CUR)

		var box Box
		err := binary.Read(file, binary.BigEndian, &box.Size)
		if err != nil {
			fmt.Println("Error reading size in parseMinf:", err)
			break
		}
		err = binary.Read(file, binary.BigEndian, &box.Type)
		if err != nil {
			fmt.Println("Error reading type in parseMinf:", err)
			break
		}

		boxType := string(box.Type[:])
		fmt.Printf("      %s [size:%d, offset:0x%X]\n", boxType, box.Size, offset)

		switch boxType {
		case "vmhd":
			fmt.Println("        Video Media Header (vmhd) found")
			parseVmhd(file, box.Size-8)
		case "smhd":
			fmt.Println("        Sound Media Header (smhd) found")
			parseSmhd(file, box.Size-8)
		case "dinf":
			fmt.Println("        Data Information (dinf) found")
			// dinf는 여기서 더 파싱하지 않음 (필요 시 추가)
			_, err := file.Seek(int64(box.Size-8), os.SEEK_CUR)
			if err != nil {
				fmt.Println("Seek error in parseMinf (dinf):", err)
				break
			}
		case "stbl":
			fmt.Println("        Sample Table (stbl) found")
			parseStbl(file, box.Size-8)
		default:
			_, err := file.Seek(int64(box.Size-8), os.SEEK_CUR)
			if err != nil {
				fmt.Println("Seek error in parseMinf:", err)
				break
			}
		}

		parsedSize += box.Size
	}

	// 남은 데이터가 있으면 건너뛰기
	if parsedSize < remainingSize {
		_, err := file.Seek(int64(remainingSize-parsedSize), os.SEEK_CUR)
		if err != nil {
			fmt.Println("Seek error in parseMinf (remaining):", err)
		}
	}
}

// vmhd 박스 파싱
func parseVmhd(file *os.File, remainingSize uint32) {
	/*
	 * vmhd box
	 * version(1 Byte)
	 * flags(3 Byte)
	 * graphicsmode(2 Byte)
	 * opcolor(6 Byte)
	 */
	var version uint8
	var flags [3]byte
	var graphicsMode uint16
	var opcolor [3]uint16

	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseVmhd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseVmhd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &graphicsMode); err != nil {
		fmt.Println("Error reading graphicsMode in parseVmhd:", err)
		return
	}
	for i := 0; i < 3; i++ {
		if err := binary.Read(file, binary.BigEndian, &opcolor[i]); err != nil {
			fmt.Println("Error reading opcolor in parseVmhd:", err)
			return
		}
	}

	fmt.Println("vmhd:")
	fmt.Printf("  graphicsmode: %d\n", graphicsMode)
	fmt.Printf("  opcolor: %d, %d, %d\n", opcolor[0], opcolor[1], opcolor[2])
}

// smhd 박스 파싱 (Sound Media Header)
func parseSmhd(file *os.File, remainingSize uint32) {
	/*
	 * smhd box
	 * version(1 Byte)
	 * flags(3 Byte)
	 * balance(2 Byte)
	 * reserved(2 Byte)
	 */
	var version uint8
	var flags [3]byte
	var balance int16
	var reserved uint16

	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseSmhd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseSmhd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &balance); err != nil {
		fmt.Println("Error reading balance in parseSmhd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &reserved); err != nil {
		fmt.Println("Error reading reserved in parseSmhd:", err)
		return
	}

	fmt.Println("smhd:")
	fmt.Printf("  balance: %d\n", balance)
	fmt.Printf("  reserved: %d\n", reserved)
}

// stbl 박스와 하위 박스 파싱
func parseStbl(file *os.File, remainingSize uint32) {
	parsedSize := uint32(0)

	for parsedSize < remainingSize {
		// 현재 오프셋 확인
		offset, _ := file.Seek(0, os.SEEK_CUR)

		var box Box
		err := binary.Read(file, binary.BigEndian, &box.Size)
		if err != nil {
			fmt.Println("Error reading size in parseStbl:", err)
			break
		}
		err = binary.Read(file, binary.BigEndian, &box.Type)
		if err != nil {
			fmt.Println("Error reading type in parseStbl:", err)
			break
		}

		boxType := string(box.Type[:])
		fmt.Printf("parseStbl        %s [size:%d, offset:0x%X]\n", boxType, box.Size, offset)

		switch boxType {
		case "stsd":
			fmt.Println("          Sample Description (stsd) found")
			parseStsd(file, box.Size-8)
		case "stts":
			fmt.Println("          Time-to-Sample (stts) found")
			parseStts(file, box.Size-8)
		case "ctts":
			fmt.Println("          Composition Time-to-Sample (ctts) found")
			parseCtts(file, box.Size-8)
		case "stsc":
			fmt.Println("          Sample-to-Chunk (stsc) found")
			parseStsc(file, box.Size-8)
		case "stsz":
			fmt.Println("          Sample Size (stsz) found")
			parseStsz(file, box.Size-8)
		case "stco":
			fmt.Println("          Chunk Offset (stco) found")
			parseStco(file, box.Size-8)
		default:
			_, err := file.Seek(int64(box.Size-8), os.SEEK_CUR)
			if err != nil {
				fmt.Println("Seek error in parseStbl:", err)
				break
			}
		}

		parsedSize += box.Size
	}

	// 남은 데이터가 있으면 건너뛰기
	if parsedSize < remainingSize {
		_, err := file.Seek(int64(remainingSize-parsedSize), os.SEEK_CUR)
		if err != nil {
			fmt.Println("Seek error in parseStbl (remaining):", err)
		}
	}
}

// stsd 박스 파싱
func parseStsd(file *os.File, remainingSize uint32) {
	/*
	 * stsd box
	 * version(1 Byte)
	 * flags(3 Byte)
	 * entry_count(4 Byte)
	 * entries(가변)
	 */
	var version uint8
	var flags [3]byte
	var entryCount uint32
	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseStsd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseStsd:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &entryCount); err != nil {
		fmt.Println("Error reading entryCount in parseStsd:", err)
		return
	}

	fmt.Println("stsd:")
	fmt.Printf("  entry_count: %d\n", entryCount)

	parsedSize := uint32(8)
	for i := uint32(0); i < entryCount && parsedSize < remainingSize; i++ {
		// 현재 오프셋 확인
		offset, _ := file.Seek(0, os.SEEK_CUR)

		var box Box
		err := binary.Read(file, binary.BigEndian, &box.Size)
		if err != nil {
			fmt.Println("Error reading size in parseStsd:", err)
			break
		}
		err = binary.Read(file, binary.BigEndian, &box.Type)
		if err != nil {
			fmt.Println("Error reading type in parseStsd:", err)
			break
		}

		boxType := string(box.Type[:])
		fmt.Printf("parseStsd          %s [size:%d, offset:0x%X]\n", boxType, box.Size, offset)

		// avc1 (H.264 비디오) 또는 mp4a (AAC 오디오) 처리
		switch boxType {
		case "avc1":
			fmt.Println("            AVC Sample Entry (avc1) found")
			parseAvc1(file, box.Size-8)
		case "mp4a":
			fmt.Println("            AAC Sample Entry (mp4a) found")
			// mp4a는 여기서 더 파싱하지 않음 (필요 시 추가)
			_, err := file.Seek(int64(box.Size-8), os.SEEK_CUR)
			if err != nil {
				fmt.Println("Seek error in parseStsd (mp4a):", err)
				break
			}
		default:
			_, err := file.Seek(int64(box.Size-8), os.SEEK_CUR)
			if err != nil {
				fmt.Println("Seek error in parseStsd:", err)
				break
			}
		}

		parsedSize += box.Size
	}

	// 남은 데이터가 있으면 건너뛰기
	if parsedSize < remainingSize {
		_, err := file.Seek(int64(remainingSize-parsedSize), os.SEEK_CUR)
		if err != nil {
			fmt.Println("Seek error in parseStsd (remaining):", err)
		}
	}
}

// avc1 박스 파싱 (H.264 비디오)
func parseAvc1(file *os.File, remainingSize uint32) {
	/*
	 * avc1 box
	 * reserved(6 Byte)
	 * data_reference_index(2 Byte)
	 * pre_defined(16 Byte)
	 * width(2 Byte)
	 * height(2 Byte)
	 * horizresolution(4 Byte)
	 * vertresolution(4 Byte)
	 * reserved(4 Byte)
	 * frame_count(2 Byte)
	 * compressorname(32 Byte)
	 * depth(2 Byte)
	 * pre_defined(2 Byte)
	 * avcC(가변)
	 */
	_, err := file.Seek(6, os.SEEK_CUR) // reserved
	if err != nil {
		fmt.Println("Seek error in parseAvc1 (reserved):", err)
		return
	}
	var dataReferenceIndex uint16
	if err := binary.Read(file, binary.BigEndian, &dataReferenceIndex); err != nil {
		fmt.Println("Error reading dataReferenceIndex in parseAvc1:", err)
		return
	}
	_, err = file.Seek(16, os.SEEK_CUR) // pre_defined
	if err != nil {
		fmt.Println("Seek error in parseAvc1 (pre_defined):", err)
		return
	}

	var width, height uint16
	if err := binary.Read(file, binary.BigEndian, &width); err != nil {
		fmt.Println("Error reading width in parseAvc1:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &height); err != nil {
		fmt.Println("Error reading height in parseAvc1:", err)
		return
	}

	var horizResolution, vertResolution uint32
	if err := binary.Read(file, binary.BigEndian, &horizResolution); err != nil {
		fmt.Println("Error reading horizResolution in parseAvc1:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &vertResolution); err != nil {
		fmt.Println("Error reading vertResolution in parseAvc1:", err)
		return
	}

	_, err = file.Seek(4, os.SEEK_CUR) // reserved
	if err != nil {
		fmt.Println("Seek error in parseAvc1 (reserved 4 bytes):", err)
		return
	}
	var frameCount uint16
	if err := binary.Read(file, binary.BigEndian, &frameCount); err != nil {
		fmt.Println("Error reading frameCount in parseAvc1:", err)
		return
	}

	// compressorname (32바이트, 문자열)
	compressorName := make([]byte, 32)
	if err := binary.Read(file, binary.BigEndian, &compressorName); err != nil {
		fmt.Println("Error reading compressorName in parseAvc1:", err)
		return
	}

	var depth uint16
	var preDefined int16
	if err := binary.Read(file, binary.BigEndian, &depth); err != nil {
		fmt.Println("Error reading depth in parseAvc1:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &preDefined); err != nil {
		fmt.Println("Error reading preDefined in parseAvc1:", err)
		return
	}

	fmt.Println("avc1:")
	fmt.Printf("  data_reference_index: %d\n", dataReferenceIndex)
	fmt.Printf("  width: %d\n", width)
	fmt.Printf("  height: %d\n", height)
	fmt.Printf("  horizresolution: 0x%08X\n", horizResolution)
	fmt.Printf("  vertresolution: 0x%08X\n", vertResolution)
	fmt.Printf("  frame_count: %d\n", frameCount)
	fmt.Printf("  compressorname: %s\n", string(compressorName))
	fmt.Printf("  depth: %d\n", depth)
	fmt.Printf("  pre_defined: %d\n", preDefined)

	// avcC 박스 파싱 (AVC Configuration)
	parsedSize := uint32(78) // avc1 헤더 크기
	if parsedSize < remainingSize {
		// 현재 오프셋 확인
		offset, _ := file.Seek(0, os.SEEK_CUR)

		var box Box
		err := binary.Read(file, binary.BigEndian, &box.Size)
		if err != nil {
			fmt.Printf("Error reading avcC box size: %v\n", err)
			return
		}
		err = binary.Read(file, binary.BigEndian, &box.Type)
		if err != nil {
			fmt.Printf("Error reading avcC box type: %v\n", err)
			return
		}

		boxType := string(box.Type[:])
		fmt.Printf("parseAvc1            %s [size:%d, offset:0x%X]\n", boxType, box.Size, offset)

		if boxType == "avcC" {
			fmt.Println("              AVC Configuration (avcC) found")
			parseAvcC(file, box.Size-8)
			parsedSize += box.Size
		}

		// avcC 이후 남은 데이터가 있으면 건너뛰기
		if parsedSize < remainingSize {
			_, err := file.Seek(int64(remainingSize-parsedSize), os.SEEK_CUR)
			if err != nil {
				fmt.Println("Seek error in parseAvc1 (remaining):", err)
			}
		}
	}
}

// avcC 박스 파싱 (AVC Configuration)
func parseAvcC(file *os.File, remainingSize uint32) {
	/*
	 * avcC box
	 * configurationVersion(1 Byte)
	 * AVCProfileIndication(1 Byte)
	 * profile_compatibility(1 Byte)
	 * AVCLevelIndication(1 Byte)
	 * reserved(6 비트) + lengthSizeMinusOne(2 비트),
	 * reserved(3 비트) + numOfSequenceParameterSets(5 비트),
	 * SPS(가변),
	 * numOfPictureParameterSets(1 Byte),
	 * PPS(가변)
	 */
	var configurationVersion, avcProfileIndication, profileCompatibility, avcLevelIndication uint8
	if err := binary.Read(file, binary.BigEndian, &configurationVersion); err != nil {
		fmt.Println("Error reading configurationVersion in parseAvcC:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &avcProfileIndication); err != nil {
		fmt.Println("Error reading avcProfileIndication in parseAvcC:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &profileCompatibility); err != nil {
		fmt.Println("Error reading profileCompatibility in parseAvcC:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &avcLevelIndication); err != nil {
		fmt.Println("Error reading avcLevelIndication in parseAvcC:", err)
		return
	}

	var lengthSizeMinusOneByte uint8
	if err := binary.Read(file, binary.BigEndian, &lengthSizeMinusOneByte); err != nil {
		fmt.Println("Error reading lengthSizeMinusOneByte in parseAvcC:", err)
		return
	}
	lengthSizeMinusOne := lengthSizeMinusOneByte & 0x03

	var numOfSequenceParameterSetsByte uint8
	if err := binary.Read(file, binary.BigEndian, &numOfSequenceParameterSetsByte); err != nil {
		fmt.Println("Error reading numOfSequenceParameterSetsByte in parseAvcC:", err)
		return
	}
	numOfSequenceParameterSets := numOfSequenceParameterSetsByte & 0x1F

	fmt.Println("avcC:")
	fmt.Printf("  configurationVersion: %d\n", configurationVersion)
	fmt.Printf("  AVCProfileIndication: %d\n", avcProfileIndication)
	fmt.Printf("  profile_compatibility: %d\n", profileCompatibility)
	fmt.Printf("  AVCLevelIndication: %d\n", avcLevelIndication)
	fmt.Printf("  lengthSizeMinusOne: %d\n", lengthSizeMinusOne)
	fmt.Printf("  numOfSequenceParameterSets: %d\n", numOfSequenceParameterSets)

	parsedSize := uint32(6)
	for i := uint8(0); i < numOfSequenceParameterSets && parsedSize < remainingSize; i++ {
		var spsSize uint16
		if err := binary.Read(file, binary.BigEndian, &spsSize); err != nil {
			fmt.Println("Error reading spsSize in parseAvcC:", err)
			return
		}
		sps := make([]byte, spsSize)
		if err := binary.Read(file, binary.BigEndian, &sps); err != nil {
			fmt.Println("Error reading sps in parseAvcC:", err)
			return
		}
		fmt.Printf("  SPS_%d: size=%d, data=[", i, spsSize)
		for j := 0; j < int(spsSize); j++ {
			fmt.Printf("%02X", sps[j])
			if j < int(spsSize)-1 {
				fmt.Print(" ")
			}
		}
		fmt.Println("]")
		parsedSize += uint32(2 + spsSize)
	}

	if parsedSize < remainingSize {
		var numOfPictureParameterSets uint8
		if err := binary.Read(file, binary.BigEndian, &numOfPictureParameterSets); err != nil {
			fmt.Println("Error reading numOfPictureParameterSets in parseAvcC:", err)
			return
		}
		fmt.Printf("  numOfPictureParameterSets: %d\n", numOfPictureParameterSets)
		parsedSize++

		for i := uint8(0); i < numOfPictureParameterSets && parsedSize < remainingSize; i++ {
			var ppsSize uint16
			if err := binary.Read(file, binary.BigEndian, &ppsSize); err != nil {
				fmt.Println("Error reading ppsSize in parseAvcC:", err)
				return
			}
			pps := make([]byte, ppsSize)
			if err := binary.Read(file, binary.BigEndian, &pps); err != nil {
				fmt.Println("Error reading pps in parseAvcC:", err)
				return
			}
			fmt.Printf("  PPS_%d: size=%d, data=[", i, ppsSize)
			for j := 0; j < int(ppsSize); j++ {
				fmt.Printf("%02X", pps[j])
				if j < int(ppsSize)-1 {
					fmt.Print(" ")
				}
			}
			fmt.Println("]")
			parsedSize += uint32(2 + ppsSize)
		}
	}

	// 남은 데이터가 있으면 건너뛰기
	if parsedSize < remainingSize {
		_, err := file.Seek(int64(remainingSize-parsedSize), os.SEEK_CUR)
		if err != nil {
			fmt.Println("Seek error in parseAvcC (remaining):", err)
		}
	}
}

// stts 박스 파싱 (Time-to-Sample)
func parseStts(file *os.File, remainingSize uint32) {
	/*
	 * stts box
	 * version(1 Byte)
	 * flags(3 Byte)
	 * entry_count(4 Byte)
	 * entries(가변)
	 */
	var version uint8
	var flags [3]byte
	var entryCount uint32
	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseStts:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseStts:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &entryCount); err != nil {
		fmt.Println("Error reading entryCount in parseStts:", err)
		return
	}

	fmt.Println("stts:")
	fmt.Printf("  entry_count: %d\n", entryCount)

	// 로그 출력을 줄이기 위해 첫 5개와 마지막 5개만 출력
	if entryCount > 10 {
		fmt.Println("  (Showing first 5 and last 5 entries to reduce output size)")
		for i := uint32(0); i < entryCount; i++ {
			var sampleCount, sampleDelta uint32
			if err := binary.Read(file, binary.BigEndian, &sampleCount); err != nil {
				fmt.Println("Error reading sampleCount in parseStts:", err)
				return
			}
			if err := binary.Read(file, binary.BigEndian, &sampleDelta); err != nil {
				fmt.Println("Error reading sampleDelta in parseStts:", err)
				return
			}
			if i < 5 || i >= entryCount-5 {
				fmt.Printf("  entry_%d/sample_count: %d\n", i, sampleCount)
				fmt.Printf("  entry_%d/sample_delta: %d\n", i, sampleDelta)
			}
		}
	} else {
		for i := uint32(0); i < entryCount; i++ {
			var sampleCount, sampleDelta uint32
			if err := binary.Read(file, binary.BigEndian, &sampleCount); err != nil {
				fmt.Println("Error reading sampleCount in parseStts:", err)
				return
			}
			if err := binary.Read(file, binary.BigEndian, &sampleDelta); err != nil {
				fmt.Println("Error reading sampleDelta in parseStts:", err)
				return
			}
			fmt.Printf("  entry_%d/sample_count: %d\n", i, sampleCount)
			fmt.Printf("  entry_%d/sample_delta: %d\n", i, sampleDelta)
		}
	}
}

// ctts 박스 파싱 (Composition Time-to-Sample)
func parseCtts(file *os.File, remainingSize uint32) {
	/*
	 * ctts box
	 * version(1 Byte)
	 * flags(3 Byte)
	 * entry_count(4 Byte)
	 * entries(가변)
	 */
	var version uint8
	var flags [3]byte
	var entryCount uint32
	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseCtts:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseCtts:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &entryCount); err != nil {
		fmt.Println("Error reading entryCount in parseCtts:", err)
		return
	}

	fmt.Println("ctts:")
	fmt.Printf("  entry_count: %d\n", entryCount)

	// 로그 출력을 줄이기 위해 첫 5개와 마지막 5개만 출력
	if entryCount > 10 {
		fmt.Println("  (Showing first 5 and last 5 entries to reduce output size)")
		for i := uint32(0); i < entryCount; i++ {
			var sampleCount, sampleOffset uint32
			if err := binary.Read(file, binary.BigEndian, &sampleCount); err != nil {
				fmt.Println("Error reading sampleCount in parseCtts:", err)
				return
			}
			if err := binary.Read(file, binary.BigEndian, &sampleOffset); err != nil {
				fmt.Println("Error reading sampleOffset in parseCtts:", err)
				return
			}
			if i < 5 || i >= entryCount-5 {
				fmt.Printf("  entry_%d/sample_count: %d\n", i, sampleCount)
				fmt.Printf("  entry_%d/sample_offset: %d\n", i, sampleOffset)
			}
		}
	} else {
		for i := uint32(0); i < entryCount; i++ {
			var sampleCount, sampleOffset uint32
			if err := binary.Read(file, binary.BigEndian, &sampleCount); err != nil {
				fmt.Println("Error reading sampleCount in parseCtts:", err)
				return
			}
			if err := binary.Read(file, binary.BigEndian, &sampleOffset); err != nil {
				fmt.Println("Error reading sampleOffset in parseCtts:", err)
				return
			}
			fmt.Printf("  entry_%d/sample_count: %d\n", i, sampleCount)
			fmt.Printf("  entry_%d/sample_offset: %d\n", i, sampleOffset)
		}
	}
}

// stsc 박스 파싱 (Sample-to-Chunk)
func parseStsc(file *os.File, remainingSize uint32) {
	/*
	 * stsc box
	 * version(1 Byte)
	 * flags(3 Byte)
	 * entry_count(4 Byte)
	 * entries(가변)
	 */
	var version uint8
	var flags [3]byte
	var entryCount uint32
	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseStsc:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseStsc:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &entryCount); err != nil {
		fmt.Println("Error reading entryCount in parseStsc:", err)
		return
	}

	fmt.Println("stsc:")
	fmt.Printf("  entry_count: %d\n", entryCount)

	// 로그 출력을 줄이기 위해 첫 5개와 마지막 5개만 출력
	if entryCount > 10 {
		fmt.Println("  (Showing first 5 and last 5 entries to reduce output size)")
		for i := uint32(0); i < entryCount; i++ {
			var firstChunk, samplesPerChunk, sampleDescriptionIndex uint32
			if err := binary.Read(file, binary.BigEndian, &firstChunk); err != nil {
				fmt.Println("Error reading firstChunk in parseStsc:", err)
				return
			}
			if err := binary.Read(file, binary.BigEndian, &samplesPerChunk); err != nil {
				fmt.Println("Error reading samplesPerChunk in parseStsc:", err)
				return
			}
			if err := binary.Read(file, binary.BigEndian, &sampleDescriptionIndex); err != nil {
				fmt.Println("Error reading sampleDescriptionIndex in parseStsc:", err)
				return
			}
			if i < 5 || i >= entryCount-5 {
				fmt.Printf("  entry_%d/first_chunk: %d\n", i, firstChunk)
				fmt.Printf("  entry_%d/samples_per_chunk: %d\n", i, samplesPerChunk)
				fmt.Printf("  entry_%d/sample_description_index: %d\n", i, sampleDescriptionIndex)
			}
		}
	} else {
		for i := uint32(0); i < entryCount; i++ {
			var firstChunk, samplesPerChunk, sampleDescriptionIndex uint32
			if err := binary.Read(file, binary.BigEndian, &firstChunk); err != nil {
				fmt.Println("Error reading firstChunk in parseStsc:", err)
				return
			}
			if err := binary.Read(file, binary.BigEndian, &samplesPerChunk); err != nil {
				fmt.Println("Error reading samplesPerChunk in parseStsc:", err)
				return
			}
			if err := binary.Read(file, binary.BigEndian, &sampleDescriptionIndex); err != nil {
				fmt.Println("Error reading sampleDescriptionIndex in parseStsc:", err)
				return
			}
			fmt.Printf("  entry_%d/first_chunk: %d\n", i, firstChunk)
			fmt.Printf("  entry_%d/samples_per_chunk: %d\n", i, samplesPerChunk)
			fmt.Printf("  entry_%d/sample_description_index: %d\n", i, sampleDescriptionIndex)
		}
	}
}

// stsz 박스 파싱 (Sample Size)
func parseStsz(file *os.File, remainingSize uint32) {
	/*
	 * stsz box
	 * version(1바이트),
	 * flags(3바이트),
	 * sample_size(4바이트),
	 * sample_count(4바이트),
	 * entries(가변)
	 */
	var version uint8
	var flags [3]byte
	var sampleSize, sampleCount uint32
	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseStsz:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseStsz:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &sampleSize); err != nil {
		fmt.Println("Error reading sampleSize in parseStsz:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &sampleCount); err != nil {
		fmt.Println("Error reading sampleCount in parseStsz:", err)
		return
	}

	fmt.Println("stsz:")
	fmt.Printf("  sample_size: %d\n", sampleSize)
	fmt.Printf("  sample_count: %d\n", sampleCount)

	if sampleSize == 0 {
		// 로그 출력을 줄이기 위해 첫 5개와 마지막 5개만 출력
		if sampleCount > 10 {
			fmt.Println("  (Showing first 5 and last 5 entries to reduce output size)")
			for i := uint32(0); i < sampleCount; i++ {
				var entrySize uint32
				if err := binary.Read(file, binary.BigEndian, &entrySize); err != nil {
					fmt.Println("Error reading entrySize in parseStsz:", err)
					return
				}
				if i < 5 || i >= sampleCount-5 {
					fmt.Printf("  entry_%d/size: %d\n", i, entrySize)
				}
			}
		} else {
			for i := uint32(0); i < sampleCount; i++ {
				var entrySize uint32
				if err := binary.Read(file, binary.BigEndian, &entrySize); err != nil {
					fmt.Println("Error reading entrySize in parseStsz:", err)
					return
				}
				fmt.Printf("  entry_%d/size: %d\n", i, entrySize)
			}
		}
	}
}

// stco 박스 파싱 (Chunk Offset)
func parseStco(file *os.File, remainingSize uint32) {
	/*
	 * stco box
	 * version(1 Byte)
	 * flags(3 Byte)
	 * entry_count(4 Byte)
	 * chunk_offsets(가변)
	 */
	var version uint8
	var flags [3]byte
	var entryCount uint32
	if err := binary.Read(file, binary.BigEndian, &version); err != nil {
		fmt.Println("Error reading version in parseStco:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &flags); err != nil {
		fmt.Println("Error reading flags in parseStco:", err)
		return
	}
	if err := binary.Read(file, binary.BigEndian, &entryCount); err != nil {
		fmt.Println("Error reading entryCount in parseStco:", err)
		return
	}

	fmt.Println("stco:")
	fmt.Printf("  entry_count: %d\n", entryCount)

	// 로그 출력을 줄이기 위해 첫 5개와 마지막 5개만 출력
	if entryCount > 10 {
		fmt.Println("  (Showing first 5 and last 5 entries to reduce output size)")
		for i := uint32(0); i < entryCount; i++ {
			var chunkOffset uint32
			if err := binary.Read(file, binary.BigEndian, &chunkOffset); err != nil {
				fmt.Println("Error reading chunkOffset in parseStco:", err)
				return
			}
			if i < 5 || i >= entryCount-5 {
				fmt.Printf("  entry_%d/chunk_offset: 0x%X\n", i, chunkOffset)
			}
		}
	} else {
		for i := uint32(0); i < entryCount; i++ {
			var chunkOffset uint32
			if err := binary.Read(file, binary.BigEndian, &chunkOffset); err != nil {
				fmt.Println("Error reading chunkOffset in parseStco:", err)
				return
			}
			fmt.Printf("  entry_%d/chunk_offset: 0x%X\n", i, chunkOffset)
		}
	}
}
