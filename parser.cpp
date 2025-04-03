#include <cstdint>
#include <cstdio>
#include <cstring>
#include <string>
#include <vector>
#include <arpa/inet.h> // ntohl, ntohs 사용 (바이트 순서 변환)

using namespace std;

// Box 구조체: MP4 박스의 기본 구조
struct Box {
    uint32_t size;
    char type[4];
};

// 함수 선언
void parseFtyp(FILE* file, uint32_t remainingSize);
void parseMoov(FILE* file, uint32_t remainingSize);
void parseMvhd(FILE* file, uint32_t remainingSize);
void parseTrak(FILE* file, uint32_t remainingSize);
void parseTkhd(FILE* file, uint32_t remainingSize);
void parseEdts(FILE* file, uint32_t remainingSize);
void parseElst(FILE* file, uint32_t remainingSize);
void parseMdia(FILE* file, uint32_t remainingSize);
void parseMdhd(FILE* file, uint32_t remainingSize);
string parseLanguage(uint16_t lang);
void parseMinf(FILE* file, uint32_t remainingSize);
void parseVmhd(FILE* file, uint32_t remainingSize);
void parseSmhd(FILE* file, uint32_t remainingSize);
void parseStbl(FILE* file, uint32_t remainingSize);
void parseStsd(FILE* file, uint32_t remainingSize);
void parseAvc1(FILE* file, uint32_t remainingSize);
void parseAvcC(FILE* file, uint32_t remainingSize);
void parseStts(FILE* file, uint32_t remainingSize);
void parseCtts(FILE* file, uint32_t remainingSize);
void parseStsc(FILE* file, uint32_t remainingSize);
void parseStsz(FILE* file, uint32_t remainingSize);
void parseStco(FILE* file, uint32_t remainingSize);
void parseStss(FILE* file, uint32_t remainingSize);
int boolToInt(bool b);

// 메인 함수
int main(int argc, char* argv[]) {
    if (argc < 2) {
        printf("사용법: %s <mp4 파일 경로>\n", argv[0]);
        return 1;
    }

    // MP4 파일 열기
    const char* filePath = argv[1];
    FILE* file = fopen(filePath, "rb");
    if (!file) {
        printf("파일 열기 오류: %s\n", filePath);
        return 1;
    }

    // 파일을 순차적으로 읽으며 박스 파싱
    while (true) {
        // 현재 오프셋 확인
        long offset = ftell(file);

        // 박스 크기와 타입 읽기
        Box box;
        if (fread(&box.size, sizeof(uint32_t), 1, file) != 1) {
            if (feof(file)) break;
            printf("Main loop size error\n");
            break;
        }
        box.size = ntohl(box.size); // Big Endian to host
        if (fread(box.type, sizeof(char), 4, file) != 4) {
            if (feof(file)) break;
            printf("Main loop type error\n");
            break;
        }

        // 박스 타입을 문자열로 변환
        string boxType(box.type, 4);
        printf("Box: %s, Size: %u, Offset: 0x%lX\n", boxType.c_str(), box.size, offset);

        // 박스별 처리
        if (boxType == "ftyp") {
            parseFtyp(file, box.size - 8);
        } else if (boxType == "moov") {
            parseMoov(file, box.size - 8);
        } else {
            // 기타 박스는 건너뛰기
            if (fseek(file, box.size - 8, SEEK_CUR) != 0) {
                printf("Seek error in main loop\n");
                break;
            }
        }
    }

    fclose(file);
    return 0;
}

// ftyp 박스 파싱
void parseFtyp(FILE* file, uint32_t remainingSize) {
    char majorBrand[4];
    uint32_t minorVersion;
    if (fread(majorBrand, sizeof(char), 4, file) != 4) {
        printf("Error reading majorBrand\n");
        return;
    }
    if (fread(&minorVersion, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading minorVersion\n");
        return;
    }
    minorVersion = ntohl(minorVersion);

    printf("ftyp:\n");
    printf("  major_brand: %.4s\n", majorBrand);
    printf("  minor_version: %u\n", minorVersion);

    // compatible_brands는 남은 크기만큼 읽기
    remainingSize -= 8; // major_brand + minor_version
    vector<char> compatibleBrands(remainingSize);
    if (fread(compatibleBrands.data(), sizeof(char), remainingSize, file) != remainingSize) {
        printf("Error reading compatibleBrands\n");
        return;
    }
    for (size_t i = 0; i < remainingSize; i += 4) {
        printf("  compatible_brand: %.4s\n", &compatibleBrands[i]);
    }
}

// moov 박스와 하위 박스 파싱
void parseMoov(FILE* file, uint32_t remainingSize) {
    uint32_t parsedSize = 0;

    while (parsedSize < remainingSize) {
        long offset = ftell(file);
        Box box;
        if (fread(&box.size, sizeof(uint32_t), 1, file) != 1) {
            if (feof(file)) break;
            printf("error: fread size\n");
            break;
        }
        box.size = ntohl(box.size);
        if (fread(box.type, sizeof(char), 4, file) != 4) {
            if (feof(file)) break;
            printf("error: fread type\n");
            break;
        }

        string boxType(box.type, 4);
        printf("============parse Moov  %s [size:%u, offset:0x%lX]============\n", boxType.c_str(), box.size, offset);

        if (boxType == "mvhd") {
            parseMvhd(file, box.size - 8);
        } else if (boxType == "trak") {
            parseTrak(file, box.size - 8);
        } else if (boxType == "tkhd") {
            parseTkhd(file, box.size - 8);
        } else if (boxType == "edts") {
            parseEdts(file, box.size - 8);
        } else if (boxType == "mdia") {
            parseMdia(file, box.size - 8);
        } else {
            if (fseek(file, box.size - 8, SEEK_CUR) != 0) {
                printf("Seek error in parseMoov\n");
                break;
            }
        }
        parsedSize += box.size;
    }
}

// mvhd 박스 파싱
void parseMvhd(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseMvhd\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseMvhd\n");
        return;
    }

    uint32_t creationTime, modificationTime, timescale, duration;
    if (version == 1) {
        uint64_t ct, mt, dur;
        if (fread(&ct, sizeof(uint64_t), 1, file) != 1) {
            printf("Error reading creationTime in parseMvhd\n");
            return;
        }
        if (fread(&mt, sizeof(uint64_t), 1, file) != 1) {
            printf("Error reading modificationTime in parseMvhd\n");
            return;
        }
        if (fread(&timescale, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading timescale in parseMvhd\n");
            return;
        }
        if (fread(&dur, sizeof(uint64_t), 1, file) != 1) {
            printf("Error reading duration in parseMvhd\n");
            return;
        }
        ct = ntohl(ct); // Big Endian to host (64비트)
        mt = ntohl(mt);
        timescale = ntohl(timescale);
        dur = ntohl(dur);
        creationTime = (uint32_t)ct;
        modificationTime = (uint32_t)mt;
        duration = (uint32_t)dur;
    } else {
        if (fread(&creationTime, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading creationTime in parseMvhd\n");
            return;
        }
        if (fread(&modificationTime, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading modificationTime in parseMvhd\n");
            return;
        }
        if (fread(&timescale, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading timescale in parseMvhd\n");
            return;
        }
        if (fread(&duration, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading duration in parseMvhd\n");
            return;
        }
        creationTime = ntohl(creationTime);
        modificationTime = ntohl(modificationTime);
        timescale = ntohl(timescale);
        duration = ntohl(duration);
    }

    // duration(ms) 계산
    double durationMs = (double(duration) * 1000) / double(timescale);

    // 나머지 데이터는 건너뛰기
    int versionAdjust = (version == 1) ? 16 : 0; // 64비트 필드 조정
    if (fseek(file, remainingSize - 12 - versionAdjust, SEEK_CUR) != 0) {
        printf("Seek error in parseMvhd\n");
        return;
    }

    printf("mvhd:\n");
    printf("  timescale: %u\n", timescale);
    printf("  duration: %u\n", duration);
    printf("  duration(ms): %.0f\n", durationMs);
}

// trak 박스와 하위 박스 파싱
void parseTrak(FILE* file, uint32_t remainingSize) {
    uint32_t parsedSize = 0;

    while (parsedSize < remainingSize) {
        long offset = ftell(file);
        Box box;
        if (fread(&box.size, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading size in parseTrak\n");
            break;
        }
        box.size = ntohl(box.size);
        if (fread(box.type, sizeof(char), 4, file) != 4) {
            printf("Error reading type in parseTrak\n");
            break;
        }

        string boxType(box.type, 4);
        printf("    %s [size:%u, offset:0x%lX]\n", boxType.c_str(), box.size, offset);

        if (boxType == "tkhd") {
            parseTkhd(file, box.size - 8);
        } else if (boxType == "edts") {
            parseEdts(file, box.size - 8);
        } else if (boxType == "mdia") {
            parseMdia(file, box.size - 8);
        } else {
            if (fseek(file, box.size - 8, SEEK_CUR) != 0) {
                printf("Seek error in parseTrak\n");
                break;
            }
        }
        parsedSize += box.size;
    }
}

// tkhd 박스 파싱
void parseTkhd(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseTkhd\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseTkhd\n");
        return;
    }

    // flags에서 enabled 여부 확인
    bool enabled = (flags[2] & 0x01) == 0x01;

    uint32_t creationTime, modificationTime, trackID, duration;
    if (version == 1) {
        uint64_t ct, mt, dur;
        if (fread(&ct, sizeof(uint64_t), 1, file) != 1) {
            printf("Error reading creationTime in parseTkhd\n");
            return;
        }
        if (fread(&mt, sizeof(uint64_t), 1, file) != 1) {
            printf("Error reading modificationTime in parseTkhd\n");
            return;
        }
        if (fread(&trackID, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading trackID in parseTkhd\n");
            return;
        }
        if (fseek(file, 4, SEEK_CUR) != 0) {
            printf("Seek error in parseTkhd (reserved)\n");
            return;
        }
        if (fread(&dur, sizeof(uint64_t), 1, file) != 1) {
            printf("Error reading duration in parseTkhd\n");
            return;
        }
        ct = ntohl(ct);
        mt = ntohl(mt);
        trackID = ntohl(trackID);
        dur = ntohl(dur);
        creationTime = (uint32_t)ct;
        modificationTime = (uint32_t)mt;
        duration = (uint32_t)dur;
    } else {
        if (fread(&creationTime, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading creationTime in parseTkhd\n");
            return;
        }
        if (fread(&modificationTime, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading modificationTime in parseTkhd\n");
            return;
        }
        if (fread(&trackID, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading trackID in parseTkhd\n");
            return;
        }
        if (fseek(file, 4, SEEK_CUR) != 0) {
            printf("Seek error in parseTkhd (reserved)\n");
            return;
        }
        if (fread(&duration, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading duration in parseTkhd\n");
            return;
        }
        creationTime = ntohl(creationTime);
        modificationTime = ntohl(modificationTime);
        trackID = ntohl(trackID);
        duration = ntohl(duration);
    }

    if (fseek(file, 8, SEEK_CUR) != 0) {
        printf("Seek error in parseTkhd (reserved 8 bytes)\n");
        return;
    }
    uint16_t layer, alternateGroup, volume;
    if (fread(&layer, sizeof(uint16_t), 1, file) != 1) {
        printf("Error reading layer in parseTkhd\n");
        return;
    }
    if (fread(&alternateGroup, sizeof(uint16_t), 1, file) != 1) {
        printf("Error reading alternateGroup in parseTkhd\n");
        return;
    }
    if (fread(&volume, sizeof(uint16_t), 1, file) != 1) {
        printf("Error reading volume in parseTkhd\n");
        return;
    }
    layer = ntohs(layer);
    alternateGroup = ntohs(alternateGroup);
    volume = ntohs(volume);

    if (fseek(file, 2, SEEK_CUR) != 0) {
        printf("Seek error in parseTkhd (reserved 2 bytes)\n");
        return;
    }

    // matrix (36바이트)
    int32_t matrix[9];
    for (int i = 0; i < 9; i++) {
        if (fread(&matrix[i], sizeof(int32_t), 1, file) != 1) {
            printf("Error reading matrix in parseTkhd\n");
            return;
        }
        matrix[i] = ntohl(matrix[i]);
    }

    // width, height (고정 소수점 16.16 형식)
    uint32_t width, height;
    if (fread(&width, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading width in parseTkhd\n");
        return;
    }
    if (fread(&height, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading height in parseTkhd\n");
        return;
    }
    width = ntohl(width);
    height = ntohl(height);

    printf("tkhd:\n");
    printf("  enabled: %d\n", boolToInt(enabled));
    printf("  id: %u\n", trackID);
    printf("  duration: %u\n", duration);
    printf("  volume: %u\n", volume >> 8);
    printf("  layer: %u\n", layer);
    printf("  alternate_group: %u\n", alternateGroup);
    for (int i = 0; i < 9; i++) {
        double iMatrix = double(matrix[i]) / 0x10000;
        printf("  matrix_%d: %.8f, 0x%08X\n", i, iMatrix, (uint32_t)matrix[i]);
    }
    printf("  width: %u, 0x%08X\n", width >> 16, width);
    printf("  height: %u, 0x%08X\n", height >> 16, height);
}

// bool을 int로 변환
int boolToInt(bool b) {
    return b ? 1 : 0;
}

// edts 박스와 하위 박스 파싱
void parseEdts(FILE* file, uint32_t remainingSize) {
    uint32_t parsedSize = 0;

    while (parsedSize < remainingSize) {
        long offset = ftell(file);
        Box box;
        if (fread(&box.size, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading size in parseEdts\n");
            break;
        }
        box.size = ntohl(box.size);
        if (fread(box.type, sizeof(char), 4, file) != 4) {
            printf("Error reading type in parseEdts\n");
            break;
        }

        string boxType(box.type, 4);
        printf("Edts        %s [size:%u, offset:0x%lX]\n", boxType.c_str(), box.size, offset);

        if (boxType == "elst") {
            parseElst(file, box.size - 8);
        } else {
            if (fseek(file, box.size - 8, SEEK_CUR) != 0) {
                printf("Seek error in parseEdts\n");
                break;
            }
        }
        parsedSize += box.size;
    }
}

// elst 박스 파싱
void parseElst(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    uint32_t entryCount;
    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseElst\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseElst\n");
        return;
    }
    if (fread(&entryCount, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading entryCount in parseElst\n");
        return;
    }
    entryCount = ntohl(entryCount);

    printf("elst:\n");
    printf("  entry_count: %u\n", entryCount);

    for (uint32_t i = 0; i < entryCount; i++) {
        uint32_t segmentDuration, mediaTime;
        int16_t mediaRateInt, mediaRateFrac;
        if (version == 1) {
            uint64_t sd, mt;
            if (fread(&sd, sizeof(uint64_t), 1, file) != 1) {
                printf("Error reading segmentDuration in parseElst\n");
                return;
            }
            if (fread(&mt, sizeof(uint64_t), 1, file) != 1) {
                printf("Error reading mediaTime in parseElst\n");
                return;
            }
            sd = ntohl(sd);
            mt = ntohl(mt);
            segmentDuration = (uint32_t)sd;
            mediaTime = (uint32_t)mt;
        } else {
            if (fread(&segmentDuration, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading segmentDuration in parseElst\n");
                return;
            }
            if (fread(&mediaTime, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading mediaTime in parseElst\n");
                return;
            }
            segmentDuration = ntohl(segmentDuration);
            mediaTime = ntohl(mediaTime);
        }
        if (fread(&mediaRateInt, sizeof(int16_t), 1, file) != 1) {
            printf("Error reading mediaRateInt in parseElst\n");
            return;
        }
        if (fread(&mediaRateFrac, sizeof(int16_t), 1, file) != 1) {
            printf("Error reading mediaRateFrac in parseElst\n");
            return;
        }
        mediaRateInt = ntohs(mediaRateInt);
        mediaRateFrac = ntohs(mediaRateFrac);

        printf("  entry/segment_duration: %u\n", segmentDuration);
        printf("  entry/media_time: %u\n", mediaTime);
        printf("  entry/media_rate: %d\n", mediaRateInt);
    }
}

// mdia 박스와 하위 박스 파싱
void parseMdia(FILE* file, uint32_t remainingSize) {
    uint32_t parsedSize = 0;

    while (parsedSize < remainingSize) {
        Box box;
        if (fread(&box.size, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading size in parseMdia\n");
            break;
        }
        box.size = ntohl(box.size);
        if (fread(box.type, sizeof(char), 4, file) != 4) {
            printf("Error reading type in parseMdia\n");
            break;
        }

        string boxType(box.type, 4);
        if (boxType == "mdhd") {
            printf("parseMdia          Media Information (mdhd) found\n");
            parseMdhd(file, box.size - 8);
        } else if (boxType == "minf") {
            printf("parseMdia          Media Information (minf) found\n");
            parseMinf(file, box.size - 8);
        } else {
            if (fseek(file, box.size - 8, SEEK_CUR) != 0) {
                printf("Seek error in parseMdia\n");
                break;
            }
        }
        parsedSize += box.size;
    }

    if (parsedSize < remainingSize) {
        if (fseek(file, remainingSize - parsedSize, SEEK_CUR) != 0) {
            printf("Seek error in parseMdia (remaining)\n");
        }
    }
}

// mdhd 박스 파싱
void parseMdhd(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseMdhd\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseMdhd\n");
        return;
    }

    uint32_t creationTime, modificationTime, timescale, duration;
    if (version == 1) {
        uint64_t ct, mt, dur;
        if (fread(&ct, sizeof(uint64_t), 1, file) != 1) {
            printf("Error reading creationTime in parseMdhd\n");
            return;
        }
        if (fread(&mt, sizeof(uint64_t), 1, file) != 1) {
            printf("Error reading modificationTime in parseMdhd\n");
            return;
        }
        if (fread(&timescale, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading timescale in parseMdhd\n");
            return;
        }
        if (fread(&dur, sizeof(uint64_t), 1, file) != 1) {
            printf("Error reading duration in parseMdhd\n");
            return;
        }
        ct = ntohl(ct);
        mt = ntohl(mt);
        timescale = ntohl(timescale);
        dur = ntohl(dur);
        creationTime = (uint32_t)ct;
        modificationTime = (uint32_t)mt;
        duration = (uint32_t)dur;
    } else {
        if (fread(&creationTime, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading creationTime in parseMdhd\n");
            return;
        }
        if (fread(&modificationTime, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading modificationTime in parseMdhd\n");
            return;
        }
        if (fread(&timescale, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading timescale in parseMdhd\n");
            return;
        }
        if (fread(&duration, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading duration in parseMdhd\n");
            return;
        }
        creationTime = ntohl(creationTime);
        modificationTime = ntohl(modificationTime);
        timescale = ntohl(timescale);
        duration = ntohl(duration);
    }

    // duration(ms) 계산
    double durationMs = (double(duration) * 1000) / double(timescale);

    // language (2바이트, ISO-639-2/T 형식)
    uint16_t language;
    if (fread(&language, sizeof(uint16_t), 1, file) != 1) {
        printf("Error reading language in parseMdhd\n");
        return;
    }
    language = ntohs(language);
    string langStr = parseLanguage(language);

    // pre_defined (2바이트, 건너뛰기)
    if (fseek(file, 2, SEEK_CUR) != 0) {
        printf("Seek error in parseMdhd (pre_defined)\n");
        return;
    }

    printf("mdhd:\n");
    printf("  timescale: %u\n", timescale);
    printf("  duration: %u\n", duration);
    printf("  duration(ms): %.0f\n", durationMs);
    printf("  language: %s\n", langStr.c_str());
}

// language 필드를 ISO-639-2/T 형식으로 변환
string parseLanguage(uint16_t lang) {
    uint8_t char1 = (lang >> 10) & 0x1F;
    uint8_t char2 = (lang >> 5) & 0x1F;
    uint8_t char3 = lang & 0x1F;

    if (char1 == 0 && char2 == 0 && char3 == 0) {
        return "und";
    }
    char langStr[4] = {
        (char)(char1 + 0x60),
        (char)(char2 + 0x60),
        (char)(char3 + 0x60),
        '\0'
    };
    return string(langStr);
}

// minf 박스와 하위 박스 파싱
void parseMinf(FILE* file, uint32_t remainingSize) {
    uint32_t parsedSize = 0;

    while (parsedSize < remainingSize) {
        long offset = ftell(file);
        Box box;
        if (fread(&box.size, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading size in parseMinf\n");
            break;
        }
        box.size = ntohl(box.size);
        if (fread(box.type, sizeof(char), 4, file) != 4) {
            printf("Error reading type in parseMinf\n");
            break;
        }

        string boxType(box.type, 4);
        printf("      %s [size:%u, offset:0x%lX]\n", boxType.c_str(), box.size, offset);

        if (boxType == "vmhd") {
            printf("        Video Media Header (vmhd) found\n");
            parseVmhd(file, box.size - 8);
        } else if (boxType == "smhd") {
            printf("        Sound Media Header (smhd) found\n");
            parseSmhd(file, box.size - 8);
        } else if (boxType == "dinf") {
            printf("        Data Information (dinf) found\n");
            if (fseek(file, box.size - 8, SEEK_CUR) != 0) {
                printf("Seek error in parseMinf (dinf)\n");
                break;
            }
        } else if (boxType == "stbl") {
            printf("        Sample Table (stbl) found\n");
            parseStbl(file, box.size - 8);
        } else {
            if (fseek(file, box.size - 8, SEEK_CUR) != 0) {
                printf("Seek error in parseMinf\n");
                break;
            }
        }
        parsedSize += box.size;
    }

    if (parsedSize < remainingSize) {
        if (fseek(file, remainingSize - parsedSize, SEEK_CUR) != 0) {
            printf("Seek error in parseMinf (remaining)\n");
        }
    }
}

// vmhd 박스 파싱
void parseVmhd(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    uint16_t graphicsMode;
    uint16_t opcolor[3];

    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseVmhd\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseVmhd\n");
        return;
    }
    if (fread(&graphicsMode, sizeof(uint16_t), 1, file) != 1) {
        printf("Error reading graphicsMode in parseVmhd\n");
        return;
    }
    graphicsMode = ntohs(graphicsMode);
    for (int i = 0; i < 3; i++) {
        if (fread(&opcolor[i], sizeof(uint16_t), 1, file) != 1) {
            printf("Error reading opcolor in parseVmhd\n");
            return;
        }
        opcolor[i] = ntohs(opcolor[i]);
    }

    printf("vmhd:\n");
    printf("  graphicsmode: %u\n", graphicsMode);
    printf("  opcolor: %u, %u, %u\n", opcolor[0], opcolor[1], opcolor[2]);
}

// smhd 박스 파싱
void parseSmhd(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    int16_t balance;
    uint16_t reserved;

    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseSmhd\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseSmhd\n");
        return;
    }
    if (fread(&balance, sizeof(int16_t), 1, file) != 1) {
        printf("Error reading balance in parseSmhd\n");
        return;
    }
    if (fread(&reserved, sizeof(uint16_t), 1, file) != 1) {
        printf("Error reading reserved in parseSmhd\n");
        return;
    }
    balance = ntohs(balance);
    reserved = ntohs(reserved);

    printf("smhd:\n");
    printf("  balance: %d\n", balance);
    printf("  reserved: %u\n", reserved);
}

// stbl 박스와 하위 박스 파싱
void parseStbl(FILE* file, uint32_t remainingSize) {
    uint32_t parsedSize = 0;

    while (parsedSize < remainingSize) {
        long offset = ftell(file);
        Box box;
        if (fread(&box.size, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading size in parseStbl\n");
            break;
        }
        box.size = ntohl(box.size);
        if (fread(box.type, sizeof(char), 4, file) != 4) {
            printf("Error reading type in parseStbl\n");
            break;
        }

        string boxType(box.type, 4);
        printf("parseStbl        %s [size:%u, offset:0x%lX]\n", boxType.c_str(), box.size, offset);

        if (boxType == "stsd") {
            printf("          Sample Description (stsd) found\n");
            parseStsd(file, box.size - 8);
        } else if (boxType == "stts") {
            printf("          Time-to-Sample (stts) found\n");
            parseStts(file, box.size - 8);
        } else if (boxType == "ctts") {
            printf("          Composition Time-to-Sample (ctts) found\n");
            parseCtts(file, box.size - 8);
        } else if (boxType == "stsc") {
            printf("          Sample-to-Chunk (stsc) found\n");
            parseStsc(file, box.size - 8);
        } else if (boxType == "stsz") {
            printf("          Sample Size (stsz) found\n");
            parseStsz(file, box.size - 8);
        } else if (boxType == "stco") {
            printf("          Chunk Offset (stco) found\n");
            parseStco(file, box.size - 8);
        } else if (boxType == "stss") {
            printf("          Sync Sample Table (stss) found\n");
            parseStss(file, box.size - 8);
        } 
        else {
            if (fseek(file, box.size - 8, SEEK_CUR) != 0) {
                printf("Seek error in parseStbl\n");
                break;
            }
        }
        parsedSize += box.size;
    }

    if (parsedSize < remainingSize) {
        if (fseek(file, remainingSize - parsedSize, SEEK_CUR) != 0) {
            printf("Seek error in parseStbl (remaining)\n");
        }
    }
}

// stsd 박스 파싱
void parseStsd(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    uint32_t entryCount;
    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseStsd\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseStsd\n");
        return;
    }
    if (fread(&entryCount, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading entryCount in parseStsd\n");
        return;
    }
    entryCount = ntohl(entryCount);

    printf("stsd:\n");
    printf("  entry_count: %u\n", entryCount);

    uint32_t parsedSize = 8;
    for (uint32_t i = 0; i < entryCount && parsedSize < remainingSize; i++) {
        long offset = ftell(file);
        Box box;
        if (fread(&box.size, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading size in parseStsd\n");
            break;
        }
        box.size = ntohl(box.size);
        if (fread(box.type, sizeof(char), 4, file) != 4) {
            printf("Error reading type in parseStsd\n");
            break;
        }

        string boxType(box.type, 4);
        printf("parseStsd          %s [size:%u, offset:0x%lX]\n", boxType.c_str(), box.size, offset);

        if (boxType == "avc1") {
            printf("            AVC Sample Entry (avc1) found\n");
            parseAvc1(file, box.size - 8);
        } else if (boxType == "mp4a") {
            printf("            AAC Sample Entry (mp4a) found\n");
            if (fseek(file, box.size - 8, SEEK_CUR) != 0) {
                printf("Seek error in parseStsd (mp4a)\n");
                break;
            }
        } else {
            if (fseek(file, box.size - 8, SEEK_CUR) != 0) {
                printf("Seek error in parseStsd\n");
                break;
            }
        }
        parsedSize += box.size;
    }

    if (parsedSize < remainingSize) {
        if (fseek(file, remainingSize - parsedSize, SEEK_CUR) != 0) {
            printf("Seek error in parseStsd (remaining)\n");
        }
    }
}

// avc1 박스 파싱
void parseAvc1(FILE* file, uint32_t remainingSize) {
    if (fseek(file, 6, SEEK_CUR) != 0) {
        printf("Seek error in parseAvc1 (reserved)\n");
        return;
    }
    uint16_t dataReferenceIndex;
    if (fread(&dataReferenceIndex, sizeof(uint16_t), 1, file) != 1) {
        printf("Error reading dataReferenceIndex in parseAvc1\n");
        return;
    }
    dataReferenceIndex = ntohs(dataReferenceIndex);
    if (fseek(file, 16, SEEK_CUR) != 0) {
        printf("Seek error in parseAvc1 (pre_defined)\n");
        return;
    }

    uint16_t width, height;
    if (fread(&width, sizeof(uint16_t), 1, file) != 1) {
        printf("Error reading width in parseAvc1\n");
        return;
    }
    if (fread(&height, sizeof(uint16_t), 1, file) != 1) {
        printf("Error reading height in parseAvc1\n");
        return;
    }
    width = ntohs(width);
    height = ntohs(height);

    uint32_t horizResolution, vertResolution;
    if (fread(&horizResolution, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading horizResolution in parseAvc1\n");
        return;
    }
    if (fread(&vertResolution, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading vertResolution in parseAvc1\n");
        return;
    }
    horizResolution = ntohl(horizResolution);
    vertResolution = ntohl(vertResolution);

    if (fseek(file, 4, SEEK_CUR) != 0) {
        printf("Seek error in parseAvc1 (reserved 4 bytes)\n");
        return;
    }
    uint16_t frameCount;
    if (fread(&frameCount, sizeof(uint16_t), 1, file) != 1) {
        printf("Error reading frameCount in parseAvc1\n");
        return;
    }
    frameCount = ntohs(frameCount);

    char compressorName[32];
    if (fread(compressorName, sizeof(char), 32, file) != 32) {
        printf("Error reading compressorName in parseAvc1\n");
        return;
    }

    uint16_t depth;
    int16_t preDefined;
    if (fread(&depth, sizeof(uint16_t), 1, file) != 1) {
        printf("Error reading depth in parseAvc1\n");
        return;
    }
    if (fread(&preDefined, sizeof(int16_t), 1, file) != 1) {
        printf("Error reading preDefined in parseAvc1\n");
        return;
    }
    depth = ntohs(depth);
    preDefined = ntohs(preDefined);

    printf("avc1:\n");
    printf("  data_reference_index: %u\n", dataReferenceIndex);
    printf("  width: %u\n", width);
    printf("  height: %u\n", height);
    printf("  horizresolution: 0x%08X\n", horizResolution);
    printf("  vertresolution: 0x%08X\n", vertResolution);
    printf("  frame_count: %u\n", frameCount);
    printf("  compressorname: %.32s\n", compressorName);
    printf("  depth: %u\n", depth);
    printf("  pre_defined: %d\n", preDefined);

    uint32_t parsedSize = 78;
    if (parsedSize < remainingSize) {
        long offset = ftell(file);
        Box box;
        if (fread(&box.size, sizeof(uint32_t), 1, file) != 1) {
            printf("Error reading avcC box size\n");
            return;
        }
        box.size = ntohl(box.size);
        if (fread(box.type, sizeof(char), 4, file) != 4) {
            printf("Error reading avcC box type\n");
            return;
        }

        string boxType(box.type, 4);
        printf("parseAvc1            %s [size:%u, offset:0x%lX]\n", boxType.c_str(), box.size, offset);

        if (boxType == "avcC") {
            printf("              AVC Configuration (avcC) found\n");
            parseAvcC(file, box.size - 8);
            parsedSize += box.size;
        }

        if (parsedSize < remainingSize) {
            if (fseek(file, remainingSize - parsedSize, SEEK_CUR) != 0) {
                printf("Seek error in parseAvc1 (remaining)\n");
            }
        }
    }
}

// avcC 박스 파싱
void parseAvcC(FILE* file, uint32_t remainingSize) {
    uint8_t configurationVersion, avcProfileIndication, profileCompatibility, avcLevelIndication;
    if (fread(&configurationVersion, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading configurationVersion in parseAvcC\n");
        return;
    }
    if (fread(&avcProfileIndication, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading avcProfileIndication in parseAvcC\n");
        return;
    }
    if (fread(&profileCompatibility, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading profileCompatibility in parseAvcC\n");
        return;
    }
    if (fread(&avcLevelIndication, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading avcLevelIndication in parseAvcC\n");
        return;
    }

    uint8_t lengthSizeMinusOneByte;
    if (fread(&lengthSizeMinusOneByte, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading lengthSizeMinusOneByte in parseAvcC\n");
        return;
    }
    uint8_t lengthSizeMinusOne = lengthSizeMinusOneByte & 0x03;

    uint8_t numOfSequenceParameterSetsByte;
    if (fread(&numOfSequenceParameterSetsByte, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading numOfSequenceParameterSetsByte in parseAvcC\n");
        return;
    }
    uint8_t numOfSequenceParameterSets = numOfSequenceParameterSetsByte & 0x1F;

    printf("avcC:\n");
    printf("  configurationVersion: %u\n", configurationVersion);
    printf("  AVCProfileIndication: %u\n", avcProfileIndication);
    printf("  profile_compatibility: %u\n", profileCompatibility);
    printf("  AVCLevelIndication: %u\n", avcLevelIndication);
    printf("  lengthSizeMinusOne: %u\n", lengthSizeMinusOne);
    printf("  numOfSequenceParameterSets: %u\n", numOfSequenceParameterSets);

    uint32_t parsedSize = 6;
    for (uint8_t i = 0; i < numOfSequenceParameterSets && parsedSize < remainingSize; i++) {
        uint16_t spsSize;
        if (fread(&spsSize, sizeof(uint16_t), 1, file) != 1) {
            printf("Error reading spsSize in parseAvcC\n");
            return;
        }
        spsSize = ntohs(spsSize);
        vector<uint8_t> sps(spsSize);
        if (fread(sps.data(), sizeof(uint8_t), spsSize, file) != spsSize) {
            printf("Error reading sps in parseAvcC\n");
            return;
        }
        printf("  SPS_%u: size=%u, data=[", i, spsSize);
        for (size_t j = 0; j < spsSize; j++) {
            printf("%02X", sps[j]);
            if (j < spsSize - 1) printf(" ");
        }
        printf("]\n");
        parsedSize += 2 + spsSize;
    }

    if (parsedSize < remainingSize) {
        uint8_t numOfPictureParameterSets;
        if (fread(&numOfPictureParameterSets, sizeof(uint8_t), 1, file) != 1) {
            printf("Error reading numOfPictureParameterSets in parseAvcC\n");
            return;
        }
        printf("  numOfPictureParameterSets: %u\n", numOfPictureParameterSets);
        parsedSize++;

        for (uint8_t i = 0; i < numOfPictureParameterSets && parsedSize < remainingSize; i++) {
            uint16_t ppsSize;
            if (fread(&ppsSize, sizeof(uint16_t), 1, file) != 1) {
                printf("Error reading ppsSize in parseAvcC\n");
                return;
            }
            ppsSize = ntohs(ppsSize);
            vector<uint8_t> pps(ppsSize);
            if (fread(pps.data(), sizeof(uint8_t), ppsSize, file) != ppsSize) {
                printf("Error reading pps in parseAvcC\n");
                return;
            }
            printf("  PPS_%u: size=%u, data=[", i, ppsSize);
            for (size_t j = 0; j < ppsSize; j++) {
                printf("%02X", pps[j]);
                if (j < ppsSize - 1) printf(" ");
            }
            printf("]\n");
            parsedSize += 2 + ppsSize;
        }
    }

    if (parsedSize < remainingSize) {
        if (fseek(file, remainingSize - parsedSize, SEEK_CUR) != 0) {
            printf("Seek error in parseAvcC (remaining)\n");
        }
    }
}

// stts 박스 파싱
void parseStts(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    uint32_t entryCount;
    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseStts\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseStts\n");
        return;
    }
    if (fread(&entryCount, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading entryCount in parseStts\n");
        return;
    }
    entryCount = ntohl(entryCount);

    printf("stts:\n");
    printf("  entry_count: %u\n", entryCount);

    if (entryCount > 10) {
        printf("  (Showing first 5 and last 5 entries to reduce output size)\n");
        for (uint32_t i = 0; i < entryCount; i++) {
            uint32_t sampleCount, sampleDelta;
            if (fread(&sampleCount, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleCount in parseStts\n");
                return;
            }
            if (fread(&sampleDelta, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleDelta in parseStts\n");
                return;
            }
            sampleCount = ntohl(sampleCount);
            sampleDelta = ntohl(sampleDelta);
            if (i < 5 || i >= entryCount - 5) {
                printf("  entry_%u/sample_count: %u\n", i, sampleCount);
                printf("  entry_%u/sample_delta: %u\n", i, sampleDelta);
            }
        }
    } else {
        for (uint32_t i = 0; i < entryCount; i++) {
            uint32_t sampleCount, sampleDelta;
            if (fread(&sampleCount, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleCount in parseStts\n");
                return;
            }
            if (fread(&sampleDelta, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleDelta in parseStts\n");
                return;
            }
            sampleCount = ntohl(sampleCount);
            sampleDelta = ntohl(sampleDelta);
            printf("  entry_%u/sample_count: %u\n", i, sampleCount);
            printf("  entry_%u/sample_delta: %u\n", i, sampleDelta);
        }
    }
}

// ctts 박스 파싱
void parseCtts(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    uint32_t entryCount;
    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseCtts\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseCtts\n");
        return;
    }
    if (fread(&entryCount, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading entryCount in parseCtts\n");
        return;
    }
    entryCount = ntohl(entryCount);

    printf("ctts:\n");
    printf("  entry_count: %u\n", entryCount);

    if (entryCount > 10) {
        printf("  (Showing first 5 and last 5 entries to reduce output size)\n");
        for (uint32_t i = 0; i < entryCount; i++) {
            uint32_t sampleCount, sampleOffset;
            if (fread(&sampleCount, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleCount in parseCtts\n");
                return;
            }
            if (fread(&sampleOffset, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleOffset in parseCtts\n");
                return;
            }
            sampleCount = ntohl(sampleCount);
            sampleOffset = ntohl(sampleOffset);
            if (i < 5 || i >= entryCount - 5) {
                printf("  entry_%u/sample_count: %u\n", i, sampleCount);
                printf("  entry_%u/sample_offset: %u\n", i, sampleOffset);
            }
        }
    } else {
        for (uint32_t i = 0; i < entryCount; i++) {
            uint32_t sampleCount, sampleOffset;
            if (fread(&sampleCount, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleCount in parseCtts\n");
                return;
            }
            if (fread(&sampleOffset, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleOffset in parseCtts\n");
                return;
            }
            sampleCount = ntohl(sampleCount);
            sampleOffset = ntohl(sampleOffset);
            printf("  entry_%u/sample_count: %u\n", i, sampleCount);
            printf("  entry_%u/sample_offset: %u\n", i, sampleOffset);
        }
    }
}
// stss 박스 파싱
void parseStss(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    uint32_t entryCount;
    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseStss\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseStss\n");
        return;
    }
    if (fread(&entryCount, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading entryCount in parseStss\n");
        return;
    }
    entryCount = ntohl(entryCount);

    printf("stss:\n");
    printf("  entry_count: %u\n", entryCount);

    if (entryCount > 10) {
        printf("  (Showing first 5 and last 5 entries to reduce output size)\n");
        for (uint32_t i = 0; i < entryCount; i++) {
            uint32_t sampleNumber;
            if (fread(&sampleNumber, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleNumber in parseStss\n");
                return;
            }
            sampleNumber = ntohl(sampleNumber);
            if (i < 5 || i >= entryCount - 5) {
                printf("  entry_%u/sample_number: %u\n", i, sampleNumber);
            }
        }
    } else {
        for (uint32_t i = 0; i < entryCount; i++) {
            uint32_t sampleNumber;
            if (fread(&sampleNumber, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleNumber in parseStss\n");
                return;
            }
            sampleNumber = ntohl(sampleNumber);
            printf("  entry_%u/sample_number: %u\n", i, sampleNumber);
        }
    }
}
// stsc 박스 파싱
void parseStsc(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    uint32_t entryCount;
    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseStsc\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseStsc\n");
        return;
    }
    if (fread(&entryCount, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading entryCount in parseStsc\n");
        return;
    }
    entryCount = ntohl(entryCount);

    printf("stsc:\n");
    printf("  entry_count: %u\n", entryCount);

    if (entryCount > 10) {
        printf("  (Showing first 5 and last 5 entries to reduce output size)\n");
        for (uint32_t i = 0; i < entryCount; i++) {
            uint32_t firstChunk, samplesPerChunk, sampleDescriptionIndex;
            if (fread(&firstChunk, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading firstChunk in parseStsc\n");
                return;
            }
            if (fread(&samplesPerChunk, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading samplesPerChunk in parseStsc\n");
                return;
            }
            if (fread(&sampleDescriptionIndex, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleDescriptionIndex in parseStsc\n");
                return;
            }
            firstChunk = ntohl(firstChunk);
            samplesPerChunk = ntohl(samplesPerChunk);
            sampleDescriptionIndex = ntohl(sampleDescriptionIndex);
            if (i < 5 || i >= entryCount - 5) {
                printf("  entry_%u/first_chunk: %u\n", i, firstChunk);
                printf("  entry_%u/samples_per_chunk: %u\n", i, samplesPerChunk);
                printf("  entry_%u/sample_description_index: %u\n", i, sampleDescriptionIndex);
            }
        }
    } else {
        for (uint32_t i = 0; i < entryCount; i++) {
            uint32_t firstChunk, samplesPerChunk, sampleDescriptionIndex;
            if (fread(&firstChunk, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading firstChunk in parseStsc\n");
                return;
            }
            if (fread(&samplesPerChunk, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading samplesPerChunk in parseStsc\n");
                return;
            }
            if (fread(&sampleDescriptionIndex, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading sampleDescriptionIndex in parseStsc\n");
                return;
            }
            firstChunk = ntohl(firstChunk);
            samplesPerChunk = ntohl(samplesPerChunk);
            sampleDescriptionIndex = ntohl(sampleDescriptionIndex);
            printf("  entry_%u/first_chunk: %u\n", i, firstChunk);
            printf("  entry_%u/samples_per_chunk: %u\n", i, samplesPerChunk);
            printf("  entry_%u/sample_description_index: %u\n", i, sampleDescriptionIndex);
        }
    }
}

// stsz 박스 파싱
void parseStsz(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    uint32_t sampleSize, sampleCount;
    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseStsz\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseStsz\n");
        return;
    }
    if (fread(&sampleSize, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading sampleSize in parseStsz\n");
        return;
    }
    if (fread(&sampleCount, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading sampleCount in parseStsz\n");
        return;
    }
    sampleSize = ntohl(sampleSize);
    sampleCount = ntohl(sampleCount);

    printf("stsz:\n");
    printf("  sample_size: %u\n", sampleSize);
    printf("  sample_count: %u\n", sampleCount);

    if (sampleSize == 0) {
        if (sampleCount > 10) {
            printf("  (Showing first 5 and last 5 entries to reduce output size)\n");
            for (uint32_t i = 0; i < sampleCount; i++) {
                uint32_t entrySize;
                if (fread(&entrySize, sizeof(uint32_t), 1, file) != 1) {
                    printf("Error reading entrySize in parseStsz\n");
                    return;
                }
                entrySize = ntohl(entrySize);
                if (i < 5 || i >= sampleCount - 5) {
                    printf("  entry_%u/size: %u\n", i, entrySize);
                }
            }
        } else {
            for (uint32_t i = 0; i < sampleCount; i++) {
                uint32_t entrySize;
                if (fread(&entrySize, sizeof(uint32_t), 1, file) != 1) {
                    printf("Error reading entrySize in parseStsz\n");
                    return;
                }
                entrySize = ntohl(entrySize);
                printf("  entry_%u/size: %u\n", i, entrySize);
            }
        }
    }
}

// stco 박스 파싱
void parseStco(FILE* file, uint32_t remainingSize) {
    uint8_t version;
    uint8_t flags[3];
    uint32_t entryCount;
    if (fread(&version, sizeof(uint8_t), 1, file) != 1) {
        printf("Error reading version in parseStco\n");
        return;
    }
    if (fread(flags, sizeof(uint8_t), 3, file) != 3) {
        printf("Error reading flags in parseStco\n");
        return;
    }
    if (fread(&entryCount, sizeof(uint32_t), 1, file) != 1) {
        printf("Error reading entryCount in parseStco\n");
        return;
    }
    entryCount = ntohl(entryCount);

    printf("stco:\n");
    printf("  entry_count: %u\n", entryCount);

    if (entryCount > 10) {
        printf("  (Showing first 5 and last 5 entries to reduce output size)\n");
        for (uint32_t i = 0; i < entryCount; i++) {
            uint32_t chunkOffset;
            if (fread(&chunkOffset, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading chunkOffset in parseStco\n");
                return;
            }
            chunkOffset = ntohl(chunkOffset);
            if (i < 5 || i >= entryCount - 5) {
                printf("  entry_%u/chunk_offset: 0x%X\n", i, chunkOffset);
            }
        }
    } else {
        for (uint32_t i = 0; i < entryCount; i++) {
            uint32_t chunkOffset;
            if (fread(&chunkOffset, sizeof(uint32_t), 1, file) != 1) {
                printf("Error reading chunkOffset in parseStco\n");
                return;
            }
            chunkOffset = ntohl(chunkOffset);
            printf("  entry_%u/chunk_offset: 0x%X\n", i, chunkOffset);
        }
    }
}

// 64비트 바이트 순서 변환 (ntohll 구현)
uint64_t ntohll(uint64_t value) {
    static const int num = 1;
    if (*(char*)&num == 1) { // Little Endian
        return ((uint64_t)ntohl(value & 0xFFFFFFFF) << 32) | ntohl(value >> 32);
    }
    return value; // Big Endian
}
