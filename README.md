README
======

At the moment this program is not working yet. I am in the middle of development.

XML to SPSS Converter (xml2sav)
-------------------------------

Xml2sav is a program that will convert a custom xml file to a binary SPSS file.
SPSS is a computer program used for statistical analysis (http://www.spss.com).
Formats like csv, excel or databases, when imported into SPSS, do not convey all
information about the variables. And also the number of rows or columns is often
limited. This program solves those problems. It is originally developed for a
web based questionnaire system. The questionnaire data is exported to a xml file
and can then be converted to an SPSS sav files using xml2sav.

Usage
-----

Install xml2sav using the installer. The installer will associate xml2sav to
files with the .xsav extension. To convert a file using xml2sav, just double
click on the data file, which should have an .xsav extention, and an console
window will pop up displaying messages indicating how the conversion process
went. The result should be that one or more binary SPSS files with the .sav
extension have been created.

Command Line Options
--------------------

The command line format is: xml2sav [/q] <xml-file>

Options:
  /q - Don't wait for the enter key when the program is finished.

Input format
------------

The input xml file has an spss element in the root. It can contain multiple sav
elements. Each sav element will generate an SPSS sav file. So it is possible to
put multiple sav files in one xml file. A sav element contains a dict element
describing the dictionary (columns) and multiple case elements that define the
cases (rows). An XML Schema definition can be found in the installation
directory.

Example

```xml
<?xml version="1.0" encoding="UTF-8"?>
<spss>
  <sav name="example">
    <dict>
      <var type="numeric" name="id" decimals="0" measure="scale"/>
      <var type="numeric" name="finished" decimals="0" measure="nominal">
        <label value="1">True</label>
        <label value="0">False</label>
      </var>
      <var type="datetime" name="start_time"/>
      <var type="string" name="remote_ip" measure="nominal"/>
      <var type="numeric" name="person.age" decimals="0" measure="scale" label="What is your age?" default="18"/>
      <var type="string" name="person.name" measure="nominal" label="What is your name?"/>
      <var type="numeric" name="frequency" decimals="0" measure="ordinal">
        <label value="1">Never</label>
        <label value="2">Sometimes</label>
        <label value="3">Regulary</label>
        <label value="4">Often</label>
      </var>
      <var type="date" name="person.dateofbirth" measure="scale"/>
    </dict>
    <case>
      <val name="id">16333</val>
      <val name="finished">1</val>
      <val name="remote_ip">1.2.3.4</val>
      <val name="start_time">05-Mar-2009 13:13:37</val>
      <val name="person.age">45</val>
      <val name="person.name">Test Person 1</val>
      <val name="frequency">2</val>
      <val name="person.dateofbirth">31-Jan-1971</val>      
    </case>
    <case>
      <val name="id">16334</val>
      <val name="finished">0</val>
      <val name="remote_ip">1.2.3.5</val>
      <val name="person.age">45</val>
      <val name="person.dateofbirth">4-Feb-2007</val>      
    </case>
    <case>
      <val name="id">16335</val>
      <val name="finished">0</val>
      <val name="remote_ip">1.2.3.5</val>
      <val name="person.dateofbirth">4-Feb-2007</val>      
    </case>
  </sav>
</spss>
```

When a defined variable is not set in a case, it will be marked as missing in
the resulting SPSS sav file.

Dates are in the format dd-mmm-yyyy, with the mmm being the abreviated name of
the month in english. Datetimes are of the format dd-mmm-yyyy hh:mm:ss.
