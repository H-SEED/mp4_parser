# MP4 box

    ftyp  
    moov  
      └── mvhd  
          └── trak  
              └── tkhd  
                  └── edts  
                      └── mdia  
                          └── minf  
                              └── stbl  
                                  ├── stsd  
                                  ├── stts  
                                  ├── stsc  
                                  ├── stsz  
                                  └── stco

    mdat  


1.ftyp

     ftyp  
     └── major_brand (4 Byte)  
     └── minor_version (4 Byte)  
     └── compatible_brands (Variable Byte)

2.mvhd

     mvhd  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── creation_time (4/8 Byte)  
     └── modification_time (4/8 Byte)  
     └── timescale (4 Byte)  
     └── duration (4/8 Byte)  
     └── rate (4 Byte)  
     └── volume (2 Byte)  
     └── reserved (10 Byte)  
     └── matrix (36 Byte)  
     └── pre_defined (24 Byte)  
     └── next_track_ID (4 Byte)  

3.tkhd

     tkhd  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── creation_time (4/8 Byte)  
     └── modification_time (4/8 Byte)  
     └── track_ID (4 Byte)  
     └── reserved (4 Byte)  
     └── duration (4/8 Byte)  
     └── reserved (8 Byte)  
     └── layer (2 Byte)  
     └── alternate_group (2 Byte)  
     └── volume (2 Byte)  
     └── reserved (2 Byte)  
     └── matrix (36 Byte)  
     └── width (4 Byte)  
     └── height (4 Byte)  

4.elst

     elst  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── entry_count (4 Byte)  
     └── entries (Variable Byte)  

5.mdhd

     mdhd  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── creation_time (4/8 Byte)  
     └── modification_time (4/8 Byte)  
     └── timescale (4 Byte)  
     └── duration (4/8 Byte)  
     └── language (2 Byte)  
     └── pre_defined (2 Byte)  

6.vmhd

     vmhd  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── graphicsmode (2 Byte)  
     └── opcolor (6 Byte)  

7.smhd

     smhd  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── balance (2 Byte)  
     └── reserved (2 Byte)  

8.stsd

     stsd  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── entry_count (4 Byte)  
     └── entries (Variable Byte)  

9.avc1

     avc1  
     └── reserved (6 Byte)  
     └── data_reference_index (2 Byte)  
     └── pre_defined (16 Byte)  
     └── width (2 Byte)  
     └── height (2 Byte)  
     └── horizresolution (4 Byte)  
     └── vertresolution (4 Byte)  
     └── reserved (4 Byte)  
     └── frame_count (2 Byte)  
     └── compressorname (32 Byte)  
     └── depth (2 Byte)  
     └── pre_defined (2 Byte)  
     └── avcC (Variable Byte)  

10.avcC

     avcC  
     └── configurationVersion (1 Byte)  
     └── AVCProfileIndication (1 Byte)  
     └── profile_compatibility (1 Byte)  
     └── AVCLevelIndication (1 Byte)  
     └── reserved (6 bits) + lengthSizeMinusOne (2 bits)  
     └── reserved (3 bits) + numOfSequenceParameterSets (5 bits)  
     └── SPS (Variable Byte)  
     └── numOfPictureParameterSets (1 Byte)  
     └── PPS (Variable Byte)  

11.stts

     stts  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── entry_count (4 Byte)  
     └── entries (Variable Byte)  

12.ctts

     ctts  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── entry_count (4 Byte)  
     └── entries (Variable Byte)  

13.stsc

     stsc  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── entry_count (4 Byte)  
     └── entries (Variable Byte) 

14.stsz

     stsz  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── sample_size (4 Byte)  
     └── sample_count (4 Byte)  
     └── entries (Variable Byte)  

15.stco

     stco  
     └── version (1 Byte)  
     └── flags (3 Byte)  
     └── entry_count (4 Byte)  
     └── chunk_offsets (Variable Byte)  




